package awskms

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/dcoker/biscuit/keymanager"
	"github.com/dcoker/biscuit/shared"
	"github.com/dcoker/biscuit/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type kmsGrantsCreate struct {
	name,
	granteePrincipal,
	retiringPrincipal,
	filename *string
	operations *[]string
	allNames   *bool
}

// NewKmsGrantsCreate constructs the command to create a grant.
func NewKmsGrantsCreate(c *kingpin.CmdClause) shared.Command {
	params := &kmsGrantsCreate{}
	params.name = c.Arg("name", "Name of the secret to grant access to.").Required().String()
	params.allNames = c.Flag("all-names", "If set, the grant allows the grantee to decrypt any values encrypted under "+
		"the keys that the named secret is encrypted with.").Default("false").Bool()
	params.granteePrincipal = c.Flag("grantee-principal", "The ARN that will be granted "+
		"additional privileges.").Short('g').PlaceHolder("ARN").Required().String()
	params.retiringPrincipal = c.Flag("retiring-principal", "The ARN that can retire the "+
		"grant.").Short('e').PlaceHolder("ARN").String()
	params.operations = operationsFlag(c)
	params.filename = shared.FilenameFlag(c)
	return params
}

type grantsCreatedOutput struct {
	Name string
	// Alias -> Region -> Grant
	Aliases map[string]map[string]grantDetails
}

type grantDetails struct {
	GrantID,
	GrantToken string
}

// Run runs the command.
func (w *kmsGrantsCreate) Run() error {
	database := store.NewFileStore(*w.filename)
	values, err := database.Get(*w.name)
	if err != nil {
		return err
	}
	values = values.FilterByKeyManager(keymanager.KmsLabel)

	aliases, err := resolveValuesToAliasesAndRegions(values)
	if err != nil {
		return err
	}

	granteeArn, retireeArn, err := resolveGranteeArns(*w.granteePrincipal, *w.retiringPrincipal)

	// The template from which grants in each region are created.
	createGrantInput := kms.CreateGrantInput{
		Operations:       aws.StringSlice(*w.operations),
		GranteePrincipal: &granteeArn,
	}
	if !*w.allNames {
		createGrantInput.Constraints = &kms.GrantConstraints{
			EncryptionContextSubset: map[string]*string{"SecretName": w.name},
		}
	}
	if len(retireeArn) > 0 {
		createGrantInput.RetiringPrincipal = &retireeArn
	}

	grantName, err := computeGrantName(createGrantInput)
	if err != nil {
		return err
	}
	createGrantInput.Name = aws.String(grantName)

	output := grantsCreatedOutput{
		Name:    grantName,
		Aliases: make(map[string]map[string]grantDetails),
	}
	for alias, regionList := range aliases {
		mrk, err := NewMultiRegionKey(alias, regionList, "")
		if err != nil {
			return err
		}
		results, err := mrk.AddGrant(createGrantInput)
		if err != nil {
			return err
		}
		regionToGrantDetails := make(map[string]grantDetails)
		for region, grant := range results {
			regionToGrantDetails[region] = grantDetails{
				GrantID:    *grant.GrantId,
				GrantToken: *grant.GrantToken}
		}
		output.Aliases[alias] = regionToGrantDetails
	}
	fmt.Print(shared.MustYaml(output))
	return nil
}

func computeGrantName(input kms.CreateGrantInput) (string, error) {
	callerIdentity, err := sts.New(session.New()).GetCallerIdentity(nil)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	gob.Register(kms.CreateGrantInput{})
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode([]interface{}{input, callerIdentity.Arn}); err != nil {
		panic(err)
	}
	hashed := sha1.Sum(buf.Bytes())
	return GrantPrefix + hex.EncodeToString(hashed[:])[:10], nil
}

func resolveValuesToAliasesAndRegions(values store.ValueList) (map[string][]string, error) {
	// The KeyID field may refer to a key/ or alias/ ARN. We need to resolve the alias for any key/ ARN
	// so that we can act on them across multiple regions. This loop resolves key/ ARNs into their appropriate
	// aliases, and maintains a list of regions for each alias.
	aliases := make(map[string][]string)
	for _, v := range values {
		arn, err := keymanager.NewARN(v.KeyID)
		if err != nil {
			return nil, err
		}
		if arn.IsKmsAlias() {
			aliases["alias/"+arn.Resource] = append(aliases["alias/"+arn.Resource], arn.Region)
		} else if arn.IsKmsKey() {
			region := arn.Region
			client := kmsHelper{kms.New(session.New(&aws.Config{Region: &region}))}
			alias, err := client.GetAliasByKeyID(arn.Resource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: Unable to find an alias for this key: %s\n", v.KeyID, err)
				return nil, err
			}
			aliases[alias] = append(aliases[alias], arn.Region)
		} else {
			return nil, err
		}
	}
	return aliases, nil
}

func resolveGranteeArns(granteePrincipal, retiringPrincipal string) (string, string, error) {
	stsClient := sts.New(session.New(&aws.Config{}))
	callerIdentity, err := stsClient.GetCallerIdentity(nil)
	if err != nil {
		return "", "", err
	}
	granteeArn := cleanArn(*callerIdentity.Account, granteePrincipal)
	if len(granteeArn) == 0 {
		return "", "", errors.New("grantee ARN must not be empty string")
	}
	retireeArn := cleanArn(*callerIdentity.Account, retiringPrincipal)
	return granteeArn, retireeArn, nil
}
