package aesgcm256

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

const (
	Name = "aesgcm256"
)

type aesGcm256 struct{}

func New() *aesGcm256 {
	return &aesGcm256{}
}

func (c *aesGcm256) Encrypt(key []byte, data []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	block, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, block.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := block.Seal(nil, nonce, data, nil)
	ciphertext = append(ciphertext, nonce...)
	return ciphertext, nil
}

func (c *aesGcm256) Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	block, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, err
	}
	nonce := ciphertext[len(ciphertext)-block.NonceSize():]
	plaintext, err := block.Open(nil, nonce, ciphertext[:len(ciphertext)-block.NonceSize()], nil)
	return plaintext, err
}

func (c *aesGcm256) NeedsKey() bool {
	return true
}
