package algorithms

const (
	plaintextLabel = "none"
)

func init() {
	registry[plaintextLabel] = newPlain
}

type plain struct{}

func newPlain() Algorithm {
	return &plain{}
}

func (s *plain) Encrypt(_ []byte, data []byte) ([]byte, error) {
	return data, nil
}

func (s *plain) Decrypt(_ []byte, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

func (s *plain) Label() string {
	return plaintextLabel
}

func (s *plain) NeedsKey() bool {
	return false
}
