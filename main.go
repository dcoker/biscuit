//go:generate go run cmd/support/generate/main.go

package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/aws/smithy-go"
	"github.com/dcoker/biscuit/commands"
	"github.com/dcoker/biscuit/commands/awskms"
	"github.com/dcoker/biscuit/shared"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	Version = "n/a"
)

//go:embed data/*
var fs embed.FS

func main() {
	os.Setenv("COLUMNS", "80") // hack to make --help output readable

	app := kingpin.New(shared.ProgName, mustAsset("data/usage.txt"))
	app.Version(Version)
	app.UsageTemplate(kingpin.LongHelpTemplate)
	getFlags := app.Command("get", "Read a secret.")
	putFlags := app.Command("put", "Write a secret.")
	listFlags := app.Command("list", "List secrets.")
	exportFlags := app.Command("export", "Print all secrets to stdout in plaintext YAML.")
	kmsFlags := app.Command("kms", "AWS KMS-specific operations.")
	kmsIDFlags := kmsFlags.Command("get-caller-identity", "Print the AWS credentials.")
	kmsInitFlags := kmsFlags.Command("init", mustAsset("data/kmsinit.txt"))
	kmsDeprovisionFlags := kmsFlags.Command("deprovision", "Deprovision AWS resources.")
	kmsEditKeyPolicyFlags := kmsFlags.Command("edit-key-policy", mustAsset("data/kmseditkeypolicy.txt"))
	kmsGrantsFlags := kmsFlags.Command("grants", "Manage KMS grants.")
	kmsGrantsListFlags := kmsGrantsFlags.Command("list", mustAsset("data/kmsgrantslist.txt"))
	kmsGrantsCreateFlags := kmsGrantsFlags.Command("create", mustAsset("data/kmsgrantcreate.txt"))
	kmsGrantsRetireFlags := kmsGrantsFlags.Command("retire", mustAsset("data/kmsgrantsretire.txt"))

	getCommand := commands.NewGet(getFlags)
	writeCommand := commands.NewPut(putFlags)
	listCommand := commands.NewList(listFlags)
	exportCommand := commands.NewExport(exportFlags)
	kmsIDCommand := awskms.KmsGetCallerIdentity{}
	kmsEditKeyPolicy := awskms.NewKmsEditKeyPolicy(kmsEditKeyPolicyFlags)
	kmsGrantsListCommand := awskms.NewKmsGrantsList(kmsGrantsListFlags)
	kmsGrantsCreateCommand := awskms.NewKmsGrantsCreate(kmsGrantsCreateFlags)
	kmsGrantsRetireCommand := awskms.NewKmsGrantsRetire(kmsGrantsRetireFlags)
	kmsInitCommand := awskms.NewKmsInit(kmsInitFlags, mustAsset("data/awskms-key.template"))
	kmsDeprovisionCommand := awskms.NewKmsDeprovision(kmsDeprovisionFlags)

	behavior := kingpin.MustParse(app.Parse(os.Args[1:]))
	ctx := context.Background()
	var err error
	switch behavior {
	case getFlags.FullCommand():
		err = getCommand.Run(ctx)
	case putFlags.FullCommand():
		err = writeCommand.Run(ctx)
	case listFlags.FullCommand():
		err = listCommand.Run(ctx)
	case kmsIDFlags.FullCommand():
		err = kmsIDCommand.Run(ctx)
	case kmsInitFlags.FullCommand():
		err = kmsInitCommand.Run(ctx)
	case kmsEditKeyPolicyFlags.FullCommand():
		err = kmsEditKeyPolicy.Run(ctx)
	case kmsGrantsCreateFlags.FullCommand():
		err = kmsGrantsCreateCommand.Run(ctx)
	case kmsGrantsListFlags.FullCommand():
		err = kmsGrantsListCommand.Run(ctx)
	case kmsDeprovisionFlags.FullCommand():
		err = kmsDeprovisionCommand.Run(ctx)
	case kmsGrantsRetireFlags.FullCommand():
		err = kmsGrantsRetireCommand.Run(ctx)
	case exportFlags.FullCommand():
		err = exportCommand.Run(ctx)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "MissingRegion":
				fmt.Fprintf(os.Stderr, "Hint: Check or set the AWS_REGION environment variable.\n")
			case "ExpiredTokenException":
				fmt.Fprintf(os.Stderr, "Hint: Refresh your credentials.\n")
			case "InvalidCiphertextException":
				fmt.Fprintf(os.Stderr, "Hint: key_ciphertext may be corrupted.\n")
			}
		}

		os.Exit(1)
	}
}

func mustAsset(filename string) string {
	bytes, err := fs.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
