package plain

const (
	Name = "none"
)

type plain struct{}

func New() *plain {
	return &plain{}
}

func (s *plain) Encrypt(_ []byte, data []byte) ([]byte, error) {
	return data, nil
}

func (s *plain) Decrypt(_ []byte, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

func (s *plain) NeedsKey() bool {
	return false
}
