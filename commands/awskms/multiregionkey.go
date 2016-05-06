package awskms

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

// MultiRegionKey represents a collection of KMS Keys that are operated on simultaneously.
type MultiRegionKey struct {
	aliasName,
	Policy string
	regions    []string
	regionToID map[string]string
}

type regionSpecificInfo struct {
	region,
	keyID,
	policy string
	err error
}

func (r *regionSpecificInfo) Error() string {
	return fmt.Sprintf("%s: %s", r.region, r.err)
}

// NewMultiRegionKey constructs a MultiRegionKey.
func NewMultiRegionKey(aliasName string, regions []string, forceRegion string) (*MultiRegionKey, error) {
	mrk := &MultiRegionKey{aliasName: aliasName, regions: regions, regionToID: make(map[string]string)}
	results := make(chan regionSpecificInfo, len(regions))
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			output := regionSpecificInfo{region: region}
			client := kmsHelper{kms.New(session.New(&aws.Config{Region: &region}))}
			keyID, policy, err := client.GetAliasTargetAndPolicy(aliasName)
			if err != nil {
				output.err = err
			} else {
				output.policy = policy
				output.keyID = keyID
			}
			results <- output
		}(region)
	}
	wg.Wait()
	close(results)

	var policy string
	var prevRegion string
	var errs []error
	for result := range results {
		result := result
		if result.err != nil {
			errs = append(errs, &result)
			continue
		}
		mrk.regionToID[result.region] = result.keyID
		if forceRegion == result.region {
			policy = result.policy
			continue
		}
		if policy == "" {
			prevRegion = result.region
			policy = result.policy
			continue
		}
		if result.policy != policy {
			errs = append(errs, &errPolicyMismatch{prevRegion, result.region})
		}
	}
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		return nil, errors.New("multiregionkey: errors collecting key information - check -r flag?")
	}
	mrk.Policy = policy
	return mrk, nil
}

// SetKeyPolicy sets a new Key Policy.
func (m *MultiRegionKey) SetKeyPolicy(policy string) error {
	errs := make(regionErrorCollector, len(m.regions))
	var wg sync.WaitGroup
	for _, region := range m.regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			client := kmsHelper{kms.New(session.New(&aws.Config{Region: &region}))}
			if _, err := client.PutKeyPolicy(&kms.PutKeyPolicyInput{
				KeyId:      aws.String(m.regionToID[region]),
				PolicyName: aws.String("default"),
				Policy:     &policy}); err != nil {
				errs <- regionError{Region: region, Err: err}
			}
		}(region)
	}
	wg.Wait()
	close(errs)
	return errs.Coalesce()
}

type getGrantsResults struct {
	region string
	grants []*kms.GrantListEntry
}

// GetGrantDetails returns a list of grants for each region.
func (m *MultiRegionKey) GetGrantDetails() (map[string][]*kms.GrantListEntry, error) {
	errs := make(regionErrorCollector, len(m.regions))
	allGrants := make(chan getGrantsResults, len(m.regions))
	var wg sync.WaitGroup
	for _, region := range m.regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			client := kmsHelper{kms.New(session.New(&aws.Config{Region: &region}))}
			var grants []*kms.GrantListEntry
			if err := client.ListGrantsPages(
				&kms.ListGrantsInput{KeyId: aws.String(m.regionToID[region])},
				func(p *kms.ListGrantsResponse, last bool) bool {
					for _, grant := range p.Grants {
						if grant.Name != nil && strings.HasPrefix(*grant.Name, GrantPrefix) {
							grants = append(grants, grant)
						}
					}
					return true
				}); err != nil {
				errs <- regionError{region, err}
				return
			}
			allGrants <- getGrantsResults{region, grants}
		}(region)
	}
	wg.Wait()
	close(errs)
	close(allGrants)
	if err := errs.Coalesce(); err != nil {
		return nil, err
	}

	regionGrants := make(map[string][]*kms.GrantListEntry)
	for grants := range allGrants {
		regionGrants[grants.region] = grants.grants
	}

	return regionGrants, nil
}

type addGrantResults struct {
	region            string
	createGrantOutput kms.CreateGrantOutput
	err               error
}

// AddGrant adds a grant to all of the underlying regions. Returns a map of region -> grant.
func (m *MultiRegionKey) AddGrant(grant kms.CreateGrantInput) (map[string]kms.CreateGrantOutput, error) {
	results := make(chan addGrantResults, len(m.regions))
	var wg sync.WaitGroup
	for _, region := range m.regions {
		wg.Add(1)
		go func(region string, grant kms.CreateGrantInput) {
			defer wg.Done()
			grant.KeyId = aws.String(m.regionToID[region])
			kmsClient := kms.New(session.New(&aws.Config{Region: &region}))
			createGrantOutput, err := kmsClient.CreateGrant(&grant)
			if err != nil {
				results <- addGrantResults{region: region, err: err}
				return
			}
			results <- addGrantResults{region: region, createGrantOutput: *createGrantOutput}
		}(region, grant)
	}
	wg.Wait()
	close(results)

	output := make(map[string]kms.CreateGrantOutput)
	for result := range results {
		if result.err != nil {
			return nil, result.err
		}
		output[result.region] = result.createGrantOutput
	}

	return output, nil
}

// RetireGrant retires a grant in all regions.
func (m *MultiRegionKey) RetireGrant(name string) error {
	results := make(regionErrorCollector, len(m.regions))
	var wg sync.WaitGroup
	for _, region := range m.regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			kmsClient := kms.New(session.New(&aws.Config{Region: &region}))
			// Find GrantID in this region
			var grantID *string
			if err := kmsClient.ListGrantsPages(&kms.ListGrantsInput{
				KeyId: aws.String(m.regionToID[region])},
				func(p *kms.ListGrantsResponse, last bool) bool {
					for _, grant := range p.Grants {
						if grant.Name != nil && *grant.Name == name {
							grantID = grant.GrantId
							return false
						}
					}
					return true
				}); err != nil {
				results <- regionError{Region: region, Err: err}
				return
			}
			if grantID == nil {
				results <- regionError{Region: region, Err: errors.New("Grant not found.")}
				return
			}

			// Revoke by GrantID
			_, err := kmsClient.RevokeGrant(&kms.RevokeGrantInput{KeyId: aws.String(m.regionToID[region]),
				GrantId: grantID})
			results <- regionError{Region: region, Err: err}
		}(region)
	}
	wg.Wait()
	close(results)
	return results.Coalesce()
}
