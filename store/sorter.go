package store

import (
	"sort"
	"strings"

	"github.com/dcoker/biscuit/internal/aws/arn"
	"github.com/dcoker/biscuit/keymanager"
)

// Sort is a function that rearranges a ValueList in place.
type Sort func(v ValueList)

func nullSort(v ValueList) {}

// SortByKmsRegion returns a Sort that will order a ValueList such that the
// head of the list contains Values in the same order as the regions passed to
// this function.
func SortByKmsRegion(regions []string) Sort {
	if regions == nil || len(regions) < 1 || regions[0] == "" {
		return nullSort
	}
	ordering := makeOrdering(regions)
	fn := func(v ValueList) {
		sort.Stable(&sortable{
			vals: v,
			less: func(left, right Value) bool {
				return lessWithRegionOrdering(ordering, left, right)
			},
		})
	}
	return fn
}

// makeOrdering creates a map of region string to integer region preference.
// Ex: us-east-1,us-west-2 -> {us-east-1:2,us-west-2:1}
func makeOrdering(regions []string) map[string]int {
	ordering := map[string]int{}
	for i, region := range regions {
		ordering[region] = len(regions) - i
	}
	return ordering
}

type sortable struct {
	vals ValueList
	less func(left, right Value) bool
}

func (b sortable) Len() int {
	return len(b.vals)
}

func (b sortable) Swap(i, j int) {
	b.vals[i], b.vals[j] = b.vals[j], b.vals[i]
}

func (b sortable) Less(i, j int) bool {
	return b.less(b.vals[i], b.vals[j])
}

func lessWithRegionOrdering(ordering map[string]int, left Value, right Value) bool {
	nat := strings.Compare(left.KeyManager, right.KeyManager)
	if nat < 0 {
		return true
	}
	if nat > 0 {
		return false
	}
	if left.KeyManager != keymanager.KmsLabel {
		return false
	}
	leftKey, err := arn.New(left.KeyID)
	if err != nil {
		return false
	}
	rightKey, err := arn.New(right.KeyID)
	if err != nil {
		return false
	}
	// Regions with a higher "ordering" will move to the beginning of the
	// list. Regions not in the ordering will get the zero value and be
	// placed towards the end of the list.
	return ordering[leftKey.Region] > ordering[rightKey.Region]
}
