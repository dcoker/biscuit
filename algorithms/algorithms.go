package algorithms

import (
	"errors"
	"sort"
)

var (
	registry                = make(map[string]func() Algorithm)
	errUnsupportedAlgorithm = errors.New("algorithms: unsupported algorithm")
)

// Algorithm implementations encrypt and decrypt data.
type Algorithm interface {
	Encrypt(key []byte, data []byte) ([]byte, error)
	Decrypt(key []byte, ciphertext []byte) ([]byte, error)
	Label() string
	NeedsKey() bool
}

// New returns an Algorithm corresponding to the requested cipher.
func New(algorithm string) (Algorithm, error) {
	if constructor, present := registry[algorithm]; present {
		return constructor(), nil
	}
	return nil, errUnsupportedAlgorithm
}

// GetDefaultAlgorithm returns the default algorithm.
func GetDefaultAlgorithm() string {
	return secretBoxLabel
}

// GetAlgorithms returns a list of registered algorithms.
func GetAlgorithms() []string {
	var algos []string
	for k := range registry {
		algos = append(algos, k)
	}
	sort.Strings(algos)
	return algos
}
