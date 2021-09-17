package keymanager

import (
	"bytes"
	"context"
)

const (
	testingLabel = "testing"
)

func init() {
	registry[testingLabel] = newTestingKeyManager
}

// testingKeys is a key manager that uses a constant key. testingKeys is only to be used for
// integration testing.
type testingKeys struct{}

var (
	testingPlaintext  = bytes.Repeat([]byte{'x'}, 32)
	testingCiphertext = bytes.Repeat([]byte{'y'}, 32)
)

// NewTestingKeyManager returns a new testingKeys.
func newTestingKeyManager() KeyManager {
	return &testingKeys{}
}

// GenerateEnvelopeKey generates an EnvelopeKey under a specific KeyID.
//noinspection GoUnusedParameter
func (k *testingKeys) GenerateEnvelopeKey(_ context.Context, keyID, secretID string) (EnvelopeKey, error) {
	return EnvelopeKey{
		ResolvedID: "resolved",
		Plaintext:  testingPlaintext,
		Ciphertext: testingCiphertext,
	}, nil
}

// Decrypt decrypts the encrypted key.
//noinspection GoUnusedParameter
func (k *testingKeys) Decrypt(_ context.Context, keyID string, keyCiphertext []byte, secretID string) ([]byte, error) {
	return testingPlaintext, nil
}

// Label returns testingLabel
func (k *testingKeys) Label() string {
	return testingLabel
}
