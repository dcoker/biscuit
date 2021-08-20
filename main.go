package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/awserr"
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
	var err error
	switch behavior {
	case getFlags.FullCommand():
		err = getCommand.Run()
	case putFlags.FullCommand():
		err = writeCommand.Run()
	case listFlags.FullCommand():
		err = listCommand.Run()
	case kmsIDFlags.FullCommand():
		err = kmsIDCommand.Run()
	case kmsInitFlags.FullCommand():
		err = kmsInitCommand.Run()
	case kmsEditKeyPolicyFlags.FullCommand():
		err = kmsEditKeyPolicy.Run()
	case kmsGrantsCreateFlags.FullCommand():
		err = kmsGrantsCreateCommand.Run()
	case kmsGrantsListFlags.FullCommand():
		err = kmsGrantsListCommand.Run()
	case kmsDeprovisionFlags.FullCommand():
		err = kmsDeprovisionCommand.Run()
	case kmsGrantsRetireFlags.FullCommand():
		err = kmsGrantsRetireCommand.Run()
	case exportFlags.FullCommand():
		err = exportCommand.Run()
	}
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "%s\n", err)
	if awsErr, ok := err.(awserr.Error); ok {
		switch awsErr.Code() {
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

func mustAsset(filename string) string {
	bytes, err := fs.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
