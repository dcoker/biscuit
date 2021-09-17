package awskms

import (
	"context"

	"github.com/dcoker/biscuit/keymanager"
	"github.com/dcoker/biscuit/shared"
	"github.com/dcoker/biscuit/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type kmsGrantsRetire struct {
	filename, name, grantName *string
}

// NewKmsGrantsRetire constructs the command to retire grants.
func NewKmsGrantsRetire(c *kingpin.CmdClause) shared.Command {
	return &kmsGrantsRetire{
		filename:  shared.FilenameFlag(c),
		name:      shared.SecretNameArg(c),
		grantName: c.Flag("grant-name", "The ID of the Grant to revoke.").Required().String(),
	}
}

func (w *kmsGrantsRetire) Run(ctx context.Context) error {
	database := store.NewFileStore(*w.filename)
	values, err := database.Get(*w.name)
	if err != nil {
		return err
	}
	values = values.FilterByKeyManager(keymanager.KmsLabel)

	aliases, err := resolveValuesToAliasesAndRegions(ctx, values)
	if err != nil {
		return err
	}

	for aliasName, regions := range aliases {
		mrk, err := NewMultiRegionKey(ctx, aliasName, regions, "")
		if err != nil {
			return err
		}

		if err := mrk.RetireGrant(ctx, *w.grantName); err != nil {
			return err
		}
	}
	return nil
}
