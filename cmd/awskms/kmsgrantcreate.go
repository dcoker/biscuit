package awskms

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/dcoker/biscuit/cmd/internal/shared"
	myAWS "github.com/dcoker/biscuit/internal/aws"
	"github.com/dcoker/biscuit/internal/aws/arn"
	"github.com/dcoker/biscuit/internal/yaml"
	"github.com/dcoker/biscuit/keymanager"
	"github.com/dcoker/biscuit/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

type kmsGrantsCreate struct {
	name,
	granteePrincipal,
	retiringPrincipal,
	filename *string
	operations []types.GrantOperation
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
func (w *kmsGrantsCreate) Run(ctx context.Context) error {
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

	granteeArn, retireeArn, err := resolveGranteeArns(ctx, *w.granteePrincipal, *w.retiringPrincipal)

	// The template from which grants in each region are created.
	createGrantInput := kms.CreateGrantInput{
		Operations:       w.operations,
		GranteePrincipal: &granteeArn,
	}
	if !*w.allNames {
		createGrantInput.Constraints = &types.GrantConstraints{
			EncryptionContextSubset: map[string]string{"SecretName": *w.name},
		}
	}
	if len(retireeArn) > 0 {
		createGrantInput.RetiringPrincipal = &retireeArn
	}

	grantName, err := computeGrantName(ctx, createGrantInput)
	if err != nil {
		return err
	}
	createGrantInput.Name = aws.String(grantName)

	output := grantsCreatedOutput{
		Name:    grantName,
		Aliases: make(map[string]map[string]grantDetails),
	}
	for alias, regionList := range aliases {
		mrk, err := NewMultiRegionKey(ctx, alias, regionList, "")
		if err != nil {
			return err
		}
		results, err := mrk.AddGrant(ctx, createGrantInput)
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
	fmt.Print(yaml.ToString(output))
	return nil
}

func computeGrantName(ctx context.Context, input kms.CreateGrantInput) (string, error) {
	cfg := myAWS.MustNewConfig(ctx)
	stsClient := sts.NewFromConfig(cfg)
	callerIdentity, err := stsClient.GetCallerIdentity(ctx, nil)
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

func resolveValuesToAliasesAndRegions(ctx context.Context, values store.ValueList) (map[string][]string, error) {
	// The KeyID field may refer to a key/ or alias/ ARN. We need to resolve the alias for any key/ ARN
	// so that we can act on them across multiple regions. This loop resolves key/ ARNs into their appropriate
	// aliases, and maintains a list of regions for each alias.
	aliases := make(map[string][]string)
	for _, v := range values {
		arn, err := arn.New(v.KeyID)
		if err != nil {
			return nil, err
		}
		if arn.IsKmsAlias() {
			aliases["alias/"+arn.Resource] = append(aliases["alias/"+arn.Resource], arn.Region)
		} else if arn.IsKmsKey() {
			cfg := myAWS.MustNewConfig(ctx, config.WithRegion(arn.Region))
			client := kmsHelper{kms.NewFromConfig(cfg)}
			alias, err := client.GetAliasByKeyID(ctx, arn.Resource)
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

func resolveGranteeArns(ctx context.Context, granteePrincipal, retiringPrincipal string) (string, string, error) {
	cfg := myAWS.MustNewConfig(ctx)
	stsClient := sts.NewFromConfig(cfg)
	callerIdentity, err := stsClient.GetCallerIdentity(ctx, nil)
	if err != nil {
		return "", "", err
	}
	granteeArn := arn.Clean(*callerIdentity.Account, granteePrincipal)
	if len(granteeArn) == 0 {
		return "", "", errors.New("grantee ARN must not be empty string")
	}
	retireeArn := arn.Clean(*callerIdentity.Account, retiringPrincipal)
	return granteeArn, retireeArn, nil
}
