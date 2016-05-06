package awskms

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

// KmsGetCallerIdentity prints AWS client configuration info.
type KmsGetCallerIdentity struct{}

// Run prints the results of STS GetCallerIdentity.
func (w *KmsGetCallerIdentity) Run() error {
	session := session.New()
	credentials, err := session.Config.Credentials.Get()
	if err != nil {
		return err
	}
	fmt.Printf("# Credentials\n")
	fmt.Printf("AWS Credentials Provider: %s\n", credentials.ProviderName)
	fmt.Printf("AWS Access Key: %s\n", credentials.AccessKeyID)
	fmt.Printf("# STS GetCallerIdentity\n")
	stsClient := sts.New(session)
	getCallerIdentityOutput, err := stsClient.GetCallerIdentity(nil)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", getCallerIdentityOutput.String())
	return nil
}
