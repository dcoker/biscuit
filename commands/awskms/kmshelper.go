package awskms

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
)

type kmsHelper struct {
	*kms.KMS
}

func (k *kmsHelper) GetAliasByName(aliasName string) (*kms.AliasListEntry, error) {
	var foundAlias *kms.AliasListEntry
	err := k.ListAliasesPages(nil, func(input *kms.ListAliasesOutput, _ bool) bool {
		for _, alias := range input.Aliases {
			if *alias.AliasName == aliasName {
				foundAlias = alias
				return false
			}
		}
		return true
	})

	return foundAlias, err
}

func (k *kmsHelper) GetAliasByKeyID(keyID string) (string, error) {
	var foundAlias *kms.AliasListEntry
	err := k.ListAliasesPages(nil, func(input *kms.ListAliasesOutput, _ bool) bool {
		for _, alias := range input.Aliases {
			if strings.HasPrefix(*alias.AliasName, AliasPrefix) && *alias.TargetKeyId == keyID {
				foundAlias = alias
				return false
			}
		}
		return true
	})
	if foundAlias != nil {
		return *foundAlias.AliasName, err
	}
	return "", &errNoAliasFoundForKey{keyID}
}

func (k *kmsHelper) GetKeyPolicyByAlias(aliasName string) (string, error) {
	alias, err := k.GetAliasByName(aliasName)
	if err != nil {
		return "", err
	}
	if alias == nil {
		return "", &errAliasNotFound{aliasName}
	}
	getKeyPolicyOutput, err := k.GetKeyPolicy(&kms.GetKeyPolicyInput{KeyId: alias.TargetKeyId,
		PolicyName: aws.String("default")})
	if err != nil {
		return "", err
	}
	return *getKeyPolicyOutput.Policy, nil
}
