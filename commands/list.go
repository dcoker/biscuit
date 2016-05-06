package commands

import (
	"fmt"

	"github.com/dcoker/biscuit/shared"
	"github.com/dcoker/biscuit/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type list struct {
	filename *string
}

// NewList configures the command to list secrets.
func NewList(c *kingpin.CmdClause) shared.Command {
	return &list{filename: shared.FilenameFlag(c)}
}

// Run runs the command.
func (r *list) Run() error {
	database := store.NewFileStore(*r.filename)

	entries, err := database.GetAll()
	if err != nil {
		return err
	}
	for name := range entries {
		if name == store.KeyTemplateName {
			continue
		}
		fmt.Printf("%s\n", name)
	}
	return nil
}
