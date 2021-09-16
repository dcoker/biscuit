package store

import (
	"encoding/base64"
	"errors"
	"os"

	"gopkg.in/yaml.v2"
)

// KeyTemplateName is the name of the value that configures the default set of key settings.
const KeyTemplateName = "_keys"

var (
	errNoTemplateEntry = errors.New("Template not found. Please specify a key ID with --key-id, or add a " +
		KeyTemplateName + " entry.")

	// ErrNameNotFound is returned by Get if the named secret does not exist.
	ErrNameNotFound = errors.New("name not found")
)

// FileStore stores an EntryMap in a YAML file on local disk.
type FileStore string

// EntryMap represents the contents of the file.
type EntryMap map[string]ValueList

// ValueList represents a list of Values.
type ValueList []Value

// FilterByKeyManager returns a new ValueList consisting only of the Values corresponding to a specific key manager.
func (v ValueList) FilterByKeyManager(manager string) ValueList {
	var results ValueList
	for _, value := range []Value(v) {
		if value.KeyManager == manager {
			results = append(results, value)
		}
	}
	return results
}

// NewFileStore constructs a FileStore for a specific filename.
func NewFileStore(filename string) FileStore {
	return FileStore(filename)
}

// Get a value.
func (f FileStore) Get(name string) (ValueList, error) {
	entries, err := f.GetAll()
	if err != nil {
		return []Value{}, err
	}
	value, present := entries[name]
	if !present {
		return []Value{}, ErrNameNotFound
	}
	return value, nil
}

// Put a value.
func (f FileStore) Put(name string, values ValueList) error {
	entries, err := f.GetAll()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	entries[name] = values

	output, err := yaml.Marshal(entries)
	if err != nil {
		return err
	}

	// poor attempt at atomic file write
	tempfile := string(f) + ".tmp"
	if err := os.WriteFile(tempfile, output, 0644); err != nil {
		return err
	}
	return os.Rename(tempfile, string(f))
}

// GetAll returns all of the entries in the file.
func (f FileStore) GetAll() (EntryMap, error) {
	contents, err := os.ReadFile(string(f))
	entries := make(EntryMap)
	if err != nil {
		return entries, err
	}
	return entries, yaml.Unmarshal(contents, entries)
}

// IsProbablyNewStore returns true if an error returned by any of the methods in this package is likely to mean
// that the store simply does not exist yet.
func IsProbablyNewStore(err error) bool {
	return os.IsNotExist(err)
}

// GetKeyIds returns the keys specified by the template entry.
func (f FileStore) GetKeyIds() ([]Key, error) {
	entries, err := f.GetAll()
	if err != nil {
		return nil, err
	}
	template, present := entries[KeyTemplateName]
	if !present {
		return nil, errNoTemplateEntry
	}

	var keys []Key
	for _, entry := range template {
		keys = append(keys, entry.Key)
	}
	return keys, nil
}

// Key defines key and crypto settings for a particular value.
type Key struct {
	// KeyID is the key that a value is encrypted under. This identifies which key the
	// KeyManager should use.
	KeyID string `yaml:"key_id,omitempty"`
	// KeyManager indicates which key manager provided this key.
	KeyManager string `yaml:"key_manager,omitempty"`
	// Algorithm used for cryptographic operations.
	Algorithm string `yaml:"algorithm"`
}

// Value is one entry in the file.
type Value struct {
	// Key references the key and cryptographic settings for this Value.
	Key `yaml:",inline"`

	// KeyCiphertext is the encryption key that Ciphertext is encrypted with, but encrypted with a
	// key that only the Provider has.
	KeyCiphertext string `yaml:"key_ciphertext,omitempty"`
	// Ciphertext is the plaintext encrypted with the ephemeral key.
	Ciphertext string `yaml:"ciphertext,omitempty"`
}

// GetKeyCiphertext returns the base64-decoded encrypted key.
func (v *Value) GetKeyCiphertext() ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(v.KeyCiphertext)
	return decoded, err
}

// GetCiphertext returns the base64-decoded ciphertext.
func (v *Value) GetCiphertext() ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(v.Ciphertext)
	return decoded, err
}
