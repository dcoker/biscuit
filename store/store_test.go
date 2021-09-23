package store

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestStore_Empty(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "TestStore")
	defer mustRemove(tmpfile.Name())
	assert.NoError(t, err)
	store := NewFileStore(tmpfile.Name())
	contents, err := store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, contents, 0)
}

func TestStore_Lifecycle(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "TestStore")
	defer mustRemove(tmpfile.Name())
	assert.NoError(t, err)
	store := NewFileStore(tmpfile.Name())
	k1put := ValueList{{
		Key: Key{
			Algorithm: "plaintext",
			KeyID:     "key_id",
		},
		KeyCiphertext: "ciphertext",
		Ciphertext:    "ciphertext",
	}}
	if err := store.Put("k1", k1put); err != nil {
		assert.NoError(t, err)
	}
	entries, err := store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, k1put, entries["k1"])
	k1actual, err := store.Get("k1")
	assert.NoError(t, err)
	assert.Equal(t, entries["k1"], k1actual)

	k2put := ValueList{{
		Key: Key{
			Algorithm: "p",
			KeyID:     "k",
		},
		KeyCiphertext: "ct",
		Ciphertext:    "c",
	}}
	if err := store.Put("k2", k2put); err != nil {
		assert.NoError(t, err)
	}
	entries, err = store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, entries, 2)
	assert.Equal(t, k1put, entries["k1"])
	assert.Equal(t, k2put, entries["k2"])
}

func TestStore_fileDoesNotExist(t *testing.T) {
	store := NewFileStore("does_not_exist")
	_, err := store.GetAll()
	assert.True(t, errors.Is(err, fs.ErrNotExist))
}

func TestStore_writingCreatesFile(t *testing.T) {
	dir, err := os.MkdirTemp("", "TestStore")
	assert.NoError(t, err)
	defer mustRemoveAll(dir)
	filename := path.Join(dir, "secrets.yml")

	store := NewFileStore(filename)
	assert.NoError(t, store.Put("k1", []Value{}))
	contents, err := store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, contents, 1)
}

func mustRemove(filename string) {
	if err := os.Remove(filename); err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete: %s\n", filename)
	}
}

func mustRemoveAll(dir string) {
	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintf(os.Stderr, "failed to delete: %s\n", dir)
	}
}
