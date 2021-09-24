package awskms

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/dcoker/biscuit/internal/aws"
)

type cloudformationStack struct {
	region,
	stackName string
	params       []types.Parameter
	templateBody *string
	templateURL  *string
}

func (s *cloudformationStack) createAndWait(ctx context.Context) (map[string]string, error) {
	cfg := aws.MustNewConfig(ctx, config.WithRegion(s.region))
	cfclient := cloudformation.NewFromConfig(cfg)
	createStackInput := &cloudformation.CreateStackInput{
		StackName:    &s.stackName,
		Capabilities: []types.Capability{types.CapabilityCapabilityIam},
		OnFailure:    types.OnFailureRollback,
		Parameters:   s.params,
		TemplateBody: s.templateBody,
		TemplateURL:  s.templateURL,
	}
	createStackOutput, err := cfclient.CreateStack(ctx, createStackInput)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%s: Waiting for CloudFormation stack %s.\n", s.region, *createStackOutput.StackId)
	describeStackInput := &cloudformation.DescribeStacksInput{StackName: createStackOutput.StackId}
	waiter := cloudformation.NewStackCreateCompleteWaiter(cfclient)
	if err := waiter.Wait(ctx, describeStackInput, 2*time.Hour); err != nil {
		return nil, err
	}
	describeStackOutput, err := cfclient.DescribeStacks(ctx, &cloudformation.DescribeStacksInput{
		StackName: createStackOutput.StackId})
	if err != nil {
		return nil, err
	}
	if len(describeStackOutput.Stacks) == 0 {
		return nil, fmt.Errorf("DescribeStacks returned an empty stack list for %s.", *createStackOutput.StackId)
	}

	outputs := make(map[string]string)
	for _, output := range describeStackOutput.Stacks[0].Outputs {
		outputs[*output.OutputKey] = *output.OutputValue
	}
	return outputs, nil
}
