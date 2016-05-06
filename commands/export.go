package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/dcoker/biscuit/shared"
	"github.com/dcoker/biscuit/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type export struct {
	filename *string
}

// NewExport configures the flags for export.
func NewExport(c *kingpin.CmdClause) shared.Command {
	return &export{filename: shared.FilenameFlag(c)}
}

// Run the command.
func (r *export) Run() error {
	database := store.NewFileStore(*r.filename)
	entries, err := database.GetAll()
	if err != nil {
		return err
	}
	errs := 0
	for name, value := range entries {
		if name == store.KeyTemplateName {
			continue
		}

		for _, v := range value {
			bytes, err := decryptOneValue(v, name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to decrypt, skipping: %s\n", err)
				errs++
				continue
			}
			fmt.Print(shared.MustYaml(map[string]string{name: string(bytes)}))
			break
		}
	}
	if errs > 0 {
		return errors.New("there were errors exporting")
	}
	return nil
}
