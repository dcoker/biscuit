package awskms

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
)

type kmsHelper struct {
	*kms.Client
}

func (k *kmsHelper) GetAliasByName(ctx context.Context, aliasName string) (*types.AliasListEntry, error) {
	p := kms.NewListAliasesPaginator(k, &kms.ListAliasesInput{})
	for p.HasMorePages() {
		output, err := p.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve aliases")
		}
		for _, alias := range output.Aliases {
			if *alias.AliasName == aliasName {
				return &alias, nil

			}

		}
	}

	return nil, nil
}

func (k *kmsHelper) GetAliasByKeyID(ctx context.Context, keyID string) (string, error) {
	p := kms.NewListAliasesPaginator(k, &kms.ListAliasesInput{})
	for p.HasMorePages() {
		output, err := p.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("could not retrieve aliases")
		}
		for _, alias := range output.Aliases {
			if strings.HasPrefix(*alias.AliasName, AliasPrefix) && *alias.TargetKeyId == keyID {
				return *alias.AliasName, nil
			}

		}
	}

	return "", &errNoAliasFoundForKey{keyID}
}

func (k *kmsHelper) GetAliasTargetAndPolicy(ctx context.Context, aliasName string) (string, string, error) {
	alias, err := k.GetAliasByName(ctx, aliasName)
	if err != nil {
		return "", "", err
	}
	if alias == nil {
		return "", "", &errAliasNotFound{aliasName}
	}
	policyOutput, err := k.GetKeyPolicy(ctx, &kms.GetKeyPolicyInput{KeyId: alias.TargetKeyId,
		PolicyName: aws.String("default")})
	if err != nil {
		return "", "", err
	}
	return *alias.TargetKeyId, *policyOutput.Policy, nil
}
