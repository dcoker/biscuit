package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/dcoker/biscuit/algorithms"
	"github.com/dcoker/biscuit/keymanager"
	"github.com/dcoker/biscuit/shared"
	"github.com/dcoker/biscuit/store"
	"github.com/mattn/go-isatty"
	"gopkg.in/alecthomas/kingpin.v2"
)

type get struct {
	name           *string
	writeTo        *string
	filename       *string
	regionPriority *[]string
}

// NewGet constructs the command to decrypt an encrypted value.
func NewGet(c *kingpin.CmdClause) shared.Command {
	return &get{
		name:           shared.SecretNameArg(c),
		regionPriority: shared.AwsRegionPriorityFlag(c),
		writeTo: c.Flag("output", "Write to FILE instead of stdout.").
			PlaceHolder("FILE").
			Short('o').
			String(),
		filename: shared.FilenameFlag(c),
	}
}

// Run the command.
func (r *get) Run(ctx context.Context) error {
	database := store.NewFileStore(*r.filename)
	values, err := database.Get(*r.name)
	if err != nil {
		return err
	}
	store.SortByKmsRegion(*r.regionPriority)(values)
	// There may be multiple values, but we assume that each one represents the same contents
	// so we stop after processing just one successfully.
	var plaintext []byte
	for _, value := range values {
		plaintext, err = decryptOneValue(ctx, value, *r.name)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Warning: decryption under %s failed: %s\n",
				value.KeyManager,
				err)
			continue
		}
		break
	}
	if err != nil {
		return err
	}

	if len(*r.writeTo) > 0 {
		return os.WriteFile(*r.writeTo, plaintext, 0644)
	}

	fmt.Printf("%s", plaintext)
	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Printf("\n")
	}
	return nil
}

func decryptOneValue(ctx context.Context, value store.Value, name string) ([]byte, error) {
	algo, err := algorithms.Get(value.Algorithm)
	if err != nil {
		return []byte{}, err
	}
	var keyPlaintext []byte
	if algo.NeedsKey() {
		keyPlaintext, err = getPlaintextKeyFromManager(ctx, value, name)
		if err != nil {
			return nil, err
		}
	}
	decoded, err := value.GetCiphertext()
	if err != nil {
		return []byte{}, err
	}
	plaintext, err := algo.Decrypt(keyPlaintext, decoded)
	return plaintext, err
}

func getPlaintextKeyFromManager(ctx context.Context, value store.Value, name string) ([]byte, error) {
	keyManager, err := keymanager.New(value.KeyManager)
	if err != nil {
		return []byte{}, err
	}
	keyCiphertext, err := value.GetKeyCiphertext()
	if err != nil {
		return []byte{}, err
	}
	keyPlaintext, err := keyManager.Decrypt(ctx, value.Key.KeyID, keyCiphertext, name)
	if err != nil {
		return []byte{}, err
	}
	return keyPlaintext, nil
}
