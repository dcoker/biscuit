package strings_test

import (
	"testing"

	"github.com/dcoker/biscuit/internal/strings"
	"github.com/stretchr/testify/assert"
)

func TestFriendlyJoin(t *testing.T) {
	assert.Equal(t, "", strings.FriendlyJoin([]string{}))
	assert.Equal(t, "us-west-1", strings.FriendlyJoin([]string{"us-west-1"}))
	assert.Equal(t, "us-east-1 and us-west-2", strings.FriendlyJoin([]string{"us-west-2", "us-east-1"}))
	assert.Equal(t, "us-east-1, us-west-1 and us-west-2", strings.FriendlyJoin([]string{"us-west-2", "us-east-1",
		"us-west-1"}))
}
