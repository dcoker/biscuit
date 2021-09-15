package algorithms

import (
	"fmt"
	"sort"
)

var (
	registry = make(map[string]Algorithm)
)

// Algorithm implementations encrypt and decrypt data.
type Algorithm interface {
	Encrypt(key []byte, data []byte) ([]byte, error)
	Decrypt(key []byte, ciphertext []byte) ([]byte, error)
	NeedsKey() bool
}

// Register adds a value to the store of all algorithms
func Register(name string, a Algorithm) error {
	_, ok := registry[name]
	if ok {
		return fmt.Errorf("algorithm %v already registered", name)
	}
	registry[name] = a
	return nil
}

// GetRegisteredAlgorithmsNames returns a list of registered algorithms.
func GetRegisteredAlgorithmsNames() []string {
	var algos []string
	for k := range registry {
		algos = append(algos, k)
	}
	sort.Strings(algos)
	return algos
}

// Get allows us to retrieve an algorithm
func Get(name string) (Algorithm, error) {
	algo, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("algorithm %v not registered", name)
	}
	return algo, nil
}
