package aws

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func NewSession(r string) *session.Session {
	cfg := &aws.Config{}
	if r != "" {
		cfg.Region = aws.String(r)
	}

	endpoint := os.Getenv("AWS_ENDPOINT")
	if endpoint != "" {
		cfg.Endpoint = aws.String(endpoint)
	}

	level := os.Getenv("BISCUIT_DEBUG")
	if level == "true" {
		cfg.WithLogLevel(aws.LogDebug)
	}
	sess := session.New(cfg)
	return sess
}
