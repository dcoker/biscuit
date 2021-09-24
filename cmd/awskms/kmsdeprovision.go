package awskms

import (
	"context"
	"fmt"
	"time"

	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	myAWS "github.com/dcoker/biscuit/internal/aws"
	"github.com/dcoker/biscuit/shared"
	"gopkg.in/alecthomas/kingpin.v2"
)

type kmsDeprovision struct {
	regions     *[]string
	label       *string
	destructive *bool
}

// NewKmsDeprovision configures the flags for kmsDeprovision.
func NewKmsDeprovision(c *kingpin.CmdClause) shared.Command {
	params := &kmsDeprovision{}
	params.regions = regionsFlag(c)
	params.label = labelFlag(c)
	params.destructive = c.Flag("destructive",
		"If true, the resources for this label will actually be deleted.").Bool()
	return params
}

// Run the command.
func (w *kmsDeprovision) Run(ctx context.Context) error {
	var failure error
	var wg sync.WaitGroup
	for _, region := range *w.regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			if err := w.deprovisionOneRegion(ctx, region); err != nil {
				fmt.Fprintf(os.Stderr, "%s: error: %s\n", region, err)
				failure = err
			}
		}(region)
	}
	wg.Wait()

	if !*w.destructive {
		fmt.Printf("\nTo delete these resources, re-run this command with --destructive.\n")
	}
	return failure
}

func (w *kmsDeprovision) deprovisionOneRegion(ctx context.Context, region string) error {
	cfg := myAWS.MustNewConfig(ctx, config.WithRegion(region))
	aliasName := kmsAliasName(*w.label)
	stackName := cfStackName(*w.label)
	fmt.Printf("%s: Searching for label '%s'...\n", region, *w.label)
	kmsClient := kmsHelper{kms.NewFromConfig(cfg)}
	foundAlias, err := kmsClient.GetAliasByName(ctx, aliasName)
	if err != nil {
		return err
	}
	if foundAlias == nil {
		fmt.Printf("%s: No KMS Key Alias %s was found.\n", region, aliasName)
	} else {
		fmt.Printf("%s: Found alias %s for %s\n", region, aliasName, *foundAlias.TargetKeyId)
		if *w.destructive {
			fmt.Printf("%s: Deleting alias...\n", region)
			if _, err := kmsClient.DeleteAlias(ctx, &kms.DeleteAliasInput{AliasName: foundAlias.AliasName}); err != nil {
				return err
			}
			fmt.Printf("%s: ... alias deleted.\n", region)
		}
	}

	exists, err := checkCloudFormationStackExists(ctx, stackName, region)
	if err != nil {
		return err
	}
	if !exists {
		fmt.Printf("%s: No CloudFormation stack named %s was found.\n", region, stackName)
		return nil
	}
	fmt.Printf("%s: Found stack: %s\n", region, stackName)
	if *w.destructive {
		cfg := myAWS.MustNewConfig(ctx, config.WithRegion(region))
		cfclient := cloudformation.NewFromConfig(cfg)
		fmt.Printf("%s: Deleting CloudFormation stack. This may take a while...\n", region)
		if _, err := cfclient.DeleteStack(ctx, &cloudformation.DeleteStackInput{StackName: &stackName}); err != nil {
			return err
		}
		waiter := cloudformation.NewStackDeleteCompleteWaiter(cfclient)
		if err := waiter.Wait(ctx, &cloudformation.DescribeStacksInput{StackName: &stackName}, 2*time.Hour); err != nil {
			return err
		}
		fmt.Printf("%s: ... stack deleted.\n", region)
	}

	return nil
}
