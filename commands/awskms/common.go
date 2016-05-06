package awskms

import (
	"fmt"
	"strings"

	"github.com/dcoker/biscuit/shared"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	// AliasPrefix is the prefix of all KMS Key Aliases.
	AliasPrefix = "alias/" + shared.ProgName + "-"

	// GrantPrefix is the prefix of all KMS Grant Names.
	GrantPrefix = shared.ProgName + "-"
)

func kmsAliasName(label string) string {
	return AliasPrefix + label
}

func cfStackName(label string) string {
	return fmt.Sprintf("%s-%s", shared.ProgName, label)
}

type errAliasNotFound struct {
	aliasName string
}

func (e *errAliasNotFound) Error() string {
	return fmt.Sprintf("key alias '%s' not found", e.aliasName)
}

type errNoAliasFoundForKey struct {
	keyID string
}

func (e *errNoAliasFoundForKey) Error() string {
	return fmt.Sprintf("no alias found for key %s", e.keyID)
}

type errPolicyMismatch struct {
	leftRegion, rightRegion string
}

func (e *errPolicyMismatch) Error() string {
	return fmt.Sprintf("the policies in region %s and %s do not match.", e.leftRegion, e.rightRegion)
}

type regionError struct {
	Region string
	Err    error
}

func (r regionError) Error() string {
	return fmt.Sprintf("%s: %s", r.Region, r.Err)
}

type regionErrorCollector chan regionError

func (r *regionErrorCollector) Coalesce() error {
	for err := range *r {
		if err.Err != nil {
			return &err
		}
	}
	return nil
}

const knownAwsKmsOperations = "Decrypt,Encrypt,GenerateDataKey,GenerateDataKeyWithoutPlaintext,ReEncryptFrom," +
	"ReEncryptTo,CreateGrant,RetireGrant"

// OperationsFlag defines a flag for the list of AWS KMS operations
func operationsFlag(cc *kingpin.CmdClause) *[]string {
	name := "operations"
	operationsList := strings.Split(knownAwsKmsOperations, ",")
	fc := cc.Flag(name,
		"Comma-separated list of AWS KMS operations this grant is allowing. Options: "+
			strings.Join(operationsList, ", ")).
		Short('o').
		Default("Decrypt,RetireGrant")
	val := (&shared.CommaSeparatedList{}).RestrictTo(operationsList...).Min(1).Name(name)
	fc.SetValue(val)
	return &val.V
}

func regionsFlag(cc *kingpin.CmdClause) *[]string {
	name := "regions"
	fc := cc.Flag("regions",
		"Comma-delimited list of regions to provision keys in. If the enviroment variable BISCUIT_REGIONS "+
			"is set, it will be used as the default value.").
		Short('r').
		Default("us-east-1,us-west-1,us-west-2").
		Envar("BISCUIT_REGIONS")
	val := (&shared.CommaSeparatedList{}).Min(1).Name(name)
	fc.SetValue(val)
	return &val.V
}

// LabelFlag defines a flag for the label.
func labelFlag(cc *kingpin.CmdClause) *string {
	label := "label"
	return shared.StringFlag(cc.Flag(label,
		"Label for the keys created. This is used to uniquely identify the keys across regions. There can "+
			"be multiple labels in use within an AWS account. If the environment variable BISCUIT_LABEL "+
			"is set, it will be used as the default value.").
		Short('l').
		Default("default").
		Envar("BISCUIT_LABEL"),
		(&shared.StringValue{}).Regex("^[a-zA-Z0-9_-]+$").Name(label).Trimmed().MinLength(1).MaxLength(20))
}
