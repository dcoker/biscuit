package awskms

import (
	"context"
	"fmt"

	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/dcoker/biscuit/cmd/internal/shared"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	errNoEditorFound        = errors.New("Set your editor preference with VISUAL or EDITOR environment variables.")
	errNewPolicyIsZeroBytes = errors.New("No change: the new policy is empty.")
	errFileUnchanged        = errors.New("No change: the new policy is the same as the existing policy.")
)

type kmsEditKeyPolicy struct {
	label       *string
	regions     *[]string
	forceRegion *string
}

// NewKmsEditKeyPolicy configures the flags for kmsEditKeyPolicy.
func NewKmsEditKeyPolicy(c *kingpin.CmdClause) shared.Command {
	return &kmsEditKeyPolicy{
		label:   labelFlag(c),
		regions: regionsFlag(c),
		forceRegion: c.Flag("force-region",
			"If set, the key policies will not be checked for consistency between regions and "+
				"the editor will open with the policy from the specified region.").String(),
	}
}

// Run the command.
func (r *kmsEditKeyPolicy) Run(ctx context.Context) error {
	aliasName := kmsAliasName(*r.label)
	mrk, err := NewMultiRegionKey(ctx, aliasName, *r.regions, *r.forceRegion)
	if err != nil {
		return err
	}

	mrk.Policy, err = prettifyJSON(mrk.Policy)
	if err != nil {
		return err
	}

	newPolicy, err := launchEditor(mrk.Policy)
	if err != nil {
		return err
	}
	indentedPolicy, err := prettifyJSON(newPolicy)
	if err != nil {
		return err
	}

	if err := mrk.SetKeyPolicy(ctx, indentedPolicy); err != nil {
		return err
	}
	fmt.Printf("New policy saved.\n")
	return nil
}

func launchEditor(contents string) (string, error) {
	f, err := os.CreateTemp("", "secrets")
	if err != nil {
		return "", err
	}
	defer os.Remove(f.Name())
	if _, err := f.WriteString(contents); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}

	editor, err := findEditor()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(editor, f.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}

	bytes, err := os.ReadFile(f.Name())
	if err != nil {
		return "", err
	}
	newContents := strings.TrimSpace(string(bytes))
	if len(newContents) == 0 {
		return "", errNewPolicyIsZeroBytes
	}
	if newContents == strings.TrimSpace(contents) {
		return "", errFileUnchanged
	}
	return newContents, nil
}

func findEditor() (string, error) {
	for _, name := range []string{"VISUAL", "EDITOR"} {
		candidate := os.Getenv(name)
		if len(candidate) > 0 {
			return candidate, nil
		}
	}
	return "", errNoEditorFound
}

func prettifyJSON(content string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return "", err
	}
	indentedPolicyBytes, err := json.MarshalIndent(&v, "", "  ")
	if err != nil {
		return "", err
	}
	indentedPolicy := string(indentedPolicyBytes)
	return indentedPolicy, nil
}
