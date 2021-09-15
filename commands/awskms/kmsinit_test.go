package awskms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
