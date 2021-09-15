package algorithms_test

import (
	"bytes"
	"crypto/rand"
	"strings"
	"testing"

	"github.com/dcoker/biscuit/algorithms"
	"github.com/dcoker/biscuit/algorithms/aesgcm256"
	"github.com/dcoker/biscuit/algorithms/secretbox"
	"github.com/stretchr/testify/assert"
)

func TestAlgorithms(t *testing.T) {
	var testInputs = []string{"", " ", "a", "ab", "12345678", "123456789",
		strings.Repeat("beef", 128)}

	var key, wrongKey [32]byte
	_, err := rand.Read(key[:])
	assert.NoError(t, err)
	_, err = rand.Read(wrongKey[:])
	assert.NoError(t, err)

	err = algorithms.Register(secretbox.Name, secretbox.New())
	assert.NoError(t, err)
	err = algorithms.Register(aesgcm256.Name, aesgcm256.New())
	assert.NoError(t, err)

	algos := algorithms.GetRegisteredAlgorithmsNames()

	for _, algoName := range algos {
		algo, err := algorithms.Get(algoName)
		assert.NoError(t, err)
		for _, expected := range testInputs {
			ciphertext, err := algo.Encrypt(key[:], []byte(expected))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// attempt to decrypt with wrong key
			if _, err = algo.Decrypt(wrongKey[:], ciphertext); err == nil {
				t.Errorf("expected error but didn't get one")
			}
			// decrypt with correct key
			plaintext, err := algo.Decrypt(key[:], ciphertext)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !bytes.Equal([]byte(expected), plaintext) {
				t.Errorf("expected: [%d]%v  plaintext: [%d]%v", len(expected), expected, len(plaintext),
					plaintext)
			}
			// mutate a few bytes and verify that it fails to decrypt
			_, err = rand.Read(ciphertext[0:4])
			assert.NoError(t, err)
			_, err = algo.Decrypt(key[:], ciphertext)
			assert.Error(t, err)
		}
	}
}
