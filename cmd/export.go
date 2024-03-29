package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/dcoker/biscuit/cmd/internal/shared"
	"github.com/dcoker/biscuit/internal/yaml"
	"github.com/dcoker/biscuit/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type export struct {
	filename       *string
	regionPriority *[]string
}

// NewExport configures the flags for export.
func NewExport(c *kingpin.CmdClause) shared.Command {
	return &export{
		filename:       shared.FilenameFlag(c),
		regionPriority: shared.AwsRegionPriorityFlag(c),
	}
}

// Run the command.
func (r *export) Run(ctx context.Context) error {
	database := store.NewFileStore(*r.filename)
	entries, err := database.GetAll()
	if err != nil {
		return err
	}
	errs := 0
	for name, values := range entries {
		if name == store.KeyTemplateName {
			continue
		}

		store.SortByKmsRegion(*r.regionPriority)(values)
		for _, v := range values {
			bytes, err := decryptOneValue(ctx, v, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to decrypt, skipping: %s\n", err)
				errs++
				continue
			}
			fmt.Print(yaml.ToString(map[string]string{name: string(bytes)}))
			break
		}
	}
	if errs > 0 {
		return errors.New("there were errors exporting")
	}
	return nil
}
