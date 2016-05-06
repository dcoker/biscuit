package awskms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFriendlyJoin(t *testing.T) {
	assert.Equal(t, "", friendlyJoin([]string{}))
	assert.Equal(t, "us-west-1", friendlyJoin([]string{"us-west-1"}))
	assert.Equal(t, "us-east-1 and us-west-2", friendlyJoin([]string{"us-west-2", "us-east-1"}))
	assert.Equal(t, "us-east-1, us-west-1 and us-west-2", friendlyJoin([]string{"us-west-2", "us-east-1",
		"us-west-1"}))
}

func TestArnList(t *testing.T) {
	assert.Equal(t,
		[]string{},
		cleanArnList("1234", ""))
	assert.Equal(t,
		[]string{"arn:aws:iam::1234:user/eat", "arn:aws:iam::1234:user/plants"},
		cleanArnList("1234", "eat,plants"))
	assert.Equal(t,
		[]string{"arn:aws:iam::1234:role/mostly", "arn:aws:iam::1234:user/fruit"},
		cleanArnList("1234", "role/mostly,user/fruit"))
	assert.Equal(t,
		[]string{"arn:aws:iam::1234:user/rule", "arn:aws:iam::4321:role/dogs"},
		cleanArnList("1234", "arn:aws:iam::4321:role/dogs,rule"))
}
