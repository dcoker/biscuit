package secretbox

import (
	"crypto/rand"

	"errors"

	"golang.org/x/crypto/nacl/secretbox"
)

const (
	Name = "secretbox"
)

var (
	errUnableToDecrypt = errors.New("secretbox: unable to decrypt")
)

type secretBox struct{}

func New() *secretBox {
	return &secretBox{}
}

func (s *secretBox) Encrypt(key []byte, data []byte) ([]byte, error) {
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, err
	}
	var keyArr [32]byte
	copy(keyArr[:], key)
	return secretbox.Seal(nonce[:], data, &nonce, &keyArr), nil
}

func (s *secretBox) Decrypt(key []byte, ciphertext []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], ciphertext[:24])
	var keyArr [32]byte
	copy(keyArr[:], key)
	var out []byte
	out, ok := secretbox.Open(out[:0], ciphertext[24:], &nonce, &keyArr)
	if !ok {
		return nil, errUnableToDecrypt
	}
	return out, nil
}

func (s *secretBox) NeedsKey() bool {
	return true
}
