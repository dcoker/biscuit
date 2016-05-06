package keymanager

import (
	"fmt"
	"sort"
)

var (
	registry = make(map[string]func() KeyManager)
)

type errUnsupportedKeyManager struct {
	label string
}

func (e *errUnsupportedKeyManager) Error() string {
	return fmt.Sprintf("unsupported key manager '%s'", e.label)
}

// New returns a KeyManager of the requested type.
func New(label string) (KeyManager, error) {
	if constructor, present := registry[label]; present {
		return constructor(), nil
	}
	return nil, &errUnsupportedKeyManager{label}
}

// GetDefaultKeyManager returns the default key manager label.
func GetDefaultKeyManager() string {
	return KmsLabel
}

// GetKeyManagers returns a list of registered key managers.
func GetKeyManagers() []string {
	var collector []string
	for k := range registry {
		collector = append(collector, k)
	}
	sort.Strings(collector)
	return collector
}

// KeyManager represents a service that can generate envelope keys and provide decryption
// keys.
type KeyManager interface {
	GenerateEnvelopeKey(keyID, secretID string) (EnvelopeKey, error)
	Decrypt(keyID string, keyMetadata []byte, secretID string) ([]byte, error)
	Label() string
}

// EnvelopeKey represents the key used in envelope encryption.
type EnvelopeKey struct {
	// ResolvedID is the fully qualified key ID.
	ResolvedID string
	// Plaintext is the plaintext encryption key.
	Plaintext []byte
	// Ciphertext is the ciphertext of the encryption key, encrypted with a key that is managed
	// by the key manager..
	Ciphertext []byte
}

// GetPlaintextKey returns the Plaintext key as a byte array.
func (e *EnvelopeKey) GetPlaintextKey() *[32]byte {
	var plaintextArray [32]byte
	copy(plaintextArray[:], e.Plaintext)
	return &plaintextArray
}
