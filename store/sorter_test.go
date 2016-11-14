package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	west1 = Value{
		Key: Key{
			KeyID:      "arn:aws:kms:us-west-1:922329555442:key/8a97cd86-54c8-4964-b9b3-4d5d6ae98139",
			KeyManager: "kms",
		},
	}
	west2 = Value{
		Key: Key{
			KeyID:      "arn:aws:kms:us-west-2:922329555442:key/0f809ad7-ecd3-41a3-9d21-923195530c8a",
			KeyManager: "kms",
		},
	}
	east1 = Value{
		Key: Key{
			KeyID:      "arn:aws:kms:us-east-1:922329555442:key/0f809ad7-ecd3-41a3-9d21-923195530c8a",
			KeyManager: "kms",
		},
	}
	east1b = Value{
		Key: Key{
			KeyID:      "arn:aws:kms:us-east-1:922329555442:key/0f809ad7-ecd3-41a3-9d21-923195530c8a",
			KeyManager: "kms",
		},
	}
	other = Value{
		Key: Key{
			KeyID:      "some other kind of key",
			KeyManager: "testing",
		},
	}
)

func TestSortByKmsRegion(t *testing.T) {
	tests := []struct {
		input, expected ValueList
		regions         []string
	}{
		{
			input:    ValueList{},
			expected: ValueList{},
		},
		{
			input:    ValueList{west2},
			expected: ValueList{west2},
		},
		{
			input:    ValueList{},
			expected: ValueList{},
			regions:  []string{"us-west-2"},
		},
		{
			input:    ValueList{west1},
			expected: ValueList{west1},
			regions:  []string{"us-west-2"},
		},
		{
			input:    ValueList{west1, west2},
			expected: ValueList{west2, west1},
			regions:  []string{"us-west-2"},
		},
		{
			input:    ValueList{west1, west2, east1, other},
			expected: ValueList{west2, west1, east1, other},
			regions:  []string{"us-west-2"},
		},
		{
			input:    ValueList{west1, west2, east1, other},
			expected: ValueList{west2, east1, west1, other},
			regions:  []string{"us-west-2", "us-east-1"},
		},
		{
			input:    ValueList{west1, west2, east1, other, east1b},
			expected: ValueList{west2, east1, east1b, west1, other},
			regions:  []string{"us-west-2", "us-east-1"},
		},
		{
			input:    ValueList{other, east1b, west1},
			expected: ValueList{west1, east1b, other},
			regions:  []string{"us-west-1", "us-east-1"},
		},
		{
			input:    ValueList{west1, west2, east1, other},
			expected: ValueList{west1, west2, east1, other},
		},
	}
	for _, tc := range tests {
		lst := make(ValueList, len(tc.input))
		copy(lst, tc.input)
		SortByKmsRegion(tc.regions)(lst)
		assert.Equal(t, tc.expected, lst)
	}
}

func TestMakeOrdering(t *testing.T) {
	tests := []struct {
		inputs   []string
		expected map[string]int
	}{
		{nil, map[string]int{}},
		{[]string{""}, map[string]int{"": 1}},
		{nil, map[string]int{}},
		{[]string{"us-west-2", "us-east-1"}, map[string]int{"us-west-2": 2, "us-east-1": 1}},
		{[]string{"us-west-2", "us-east-1", "us-west-1"}, map[string]int{"us-west-1": 1, "us-east-1": 2, "us-west-2": 3}},
	}
	for _, tc := range tests {
		ordering := makeOrdering(tc.inputs)
		assert.Equal(t, ordering, tc.expected)
	}
}
