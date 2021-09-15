package secretbox_test

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	sb "github.com/dcoker/biscuit/algorithms/secretbox"
	"github.com/stretchr/testify/assert"
)

func TestSecretboxSymmetry(t *testing.T) {
	var key [32]byte
	_, err := rand.Read(key[:])
	assert.NoError(t, err)
	box := sb.New()

	nonces := make(map[string]struct{})
	for i := 0; i < 100; i++ {
		var message [4096]byte
		_, err := rand.Read(message[:])
		assert.NoError(t, err)
		ciphertext, err := box.Encrypt(key[:], message[:])

		// Test that nonces seem to be unique
		nonce := base64.StdEncoding.EncodeToString(ciphertext[:24])
		_, seen := nonces[nonce]
		assert.False(t, seen)
		nonces[nonce] = struct{}{}

		assert.NoError(t, err)
		plaintext, err := box.Decrypt(key[:], ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, message[:], plaintext)
	}
}
