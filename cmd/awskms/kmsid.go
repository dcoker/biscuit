package awskms

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sts"
	myAWS "github.com/dcoker/biscuit/internal/aws"
)

// KmsGetCallerIdentity prints AWS client configuration info.
type KmsGetCallerIdentity struct{}

// Run prints the results of STS GetCallerIdentity.
func (w *KmsGetCallerIdentity) Run(ctx context.Context) error {
	cfg := myAWS.MustNewConfig(ctx)
	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("# Credentials\n")
	fmt.Printf("AWS Credentials Provider: %s\n", credentials.Source)
	fmt.Printf("AWS Access Key: %s\n", credentials.AccessKeyID)
	fmt.Printf("# STS GetCallerIdentity\n")
	stsClient := sts.NewFromConfig(cfg)
	getCallerIdentityOutput, err := stsClient.GetCallerIdentity(ctx, nil)
	if err != nil {
		return err
	}
	fmt.Printf("\tAccount: %s\n\tARN: %s\n\tUserID: %s\n", *getCallerIdentityOutput.Account, *getCallerIdentityOutput.Arn, *getCallerIdentityOutput.UserId)
	return nil
}
