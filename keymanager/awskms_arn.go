package keymanager

import (
	"fmt"
	"strings"
)

type errInvalidArn struct {
	arn string
}

func (e *errInvalidArn) Error() string {
	return fmt.Sprintf("%s: invalid ARN", e.arn)
}

// ARN represents the ARN as documented by http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html
type ARN struct {
	Partition,
	Service,
	Region,
	AccountID,
	ResourceType,
	Resource string
	delimiter string
}

// NewARN parses s and constructs an ARN.
func NewARN(s string) (ARN, error) {
	s = strings.TrimSpace(s)
	splat := strings.Split(s, ":")
	if !strings.HasPrefix(s, "arn:") {
		return ARN{}, &errInvalidArn{s}
	}
	if len(splat) < 6 {
		return ARN{}, &errInvalidArn{s}
	}
	arn := ARN{
		Partition: splat[1],
		Service:   splat[2],
		Region:    splat[3],
		AccountID: splat[4],
	}
	if len(splat) == 7 {
		arn.ResourceType = splat[5]
		arn.Resource = splat[6]
		arn.delimiter = ":"
	} else if len(splat) == 6 {
		if strings.Contains(splat[5], "/") {
			resourceTypeResource := strings.SplitN(splat[5], "/", 2)
			arn.ResourceType = resourceTypeResource[0]
			arn.Resource = resourceTypeResource[1]
			arn.delimiter = "/"
		} else {
			arn.Resource = splat[5]
		}
	} else {
		return ARN{}, &errInvalidArn{s}
	}
	return arn, nil
}

func (a *ARN) String() string {
	arn := strings.Join([]string{"arn", a.Partition, a.Service, a.Region, a.AccountID, ""}, ":")
	if len(a.ResourceType) > 0 {
		arn += a.ResourceType + a.delimiter + a.Resource
	} else {
		arn += a.Resource
	}
	return arn
}

// IsKmsKey returns true iff the ARN represents a KMS Key.
func (a *ARN) IsKmsKey() bool {
	return a.Service == "kms" && a.ResourceType == "key"
}

// IsKmsAlias returns true iff the ARN represents a KMS Alias.
func (a *ARN) IsKmsAlias() bool {
	return a.Service == "kms" && a.ResourceType == "alias"
}
