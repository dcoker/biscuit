package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func NewConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
	if os.Getenv("BISCUIT_DEBUG") == "true" {
		optFns = append(optFns, config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody))
		optFns = append(optFns, config.WithRetryer(func() aws.Retryer { return aws.NopRetryer{} }))
	}
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to load SDK config, %w", err)
	}
	endpoint := os.Getenv("AWS_ENDPOINT")
	if endpoint != "" {
		cfg.EndpointResolver = aws.EndpointResolverFunc(awsEndpoint(endpoint))
	}
	return cfg, nil
}

func MustNewConfig(ctx context.Context, optFns ...func(*config.LoadOptions) error) aws.Config {
	cfg, err := NewConfig(ctx, optFns...)
	if err != nil {
		panic(err)
	}
	return cfg
}

func awsEndpoint(endpoint string) func(service, region string) (aws.Endpoint, error) {
	return func(service, region string) (aws.Endpoint, error) {
		return aws.Endpoint{
			SigningRegion:     region,
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	}
}
