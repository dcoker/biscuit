package keymanager

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
)

const (
	// KmsLabel is the label for the AWS KMS.
	KmsLabel = "kms"
)

func init() {
	registry[KmsLabel] = NewKms
}

// Kms is a KeyManager for AWS KMS.
type Kms struct{}

// NewKms returns a new Kms.
func NewKms() KeyManager {
	return &Kms{}
}

// GenerateEnvelopeKey generates an EnvelopeKey under a specific KeyID.
func (k *Kms) GenerateEnvelopeKey(keyID string, secretID string) (EnvelopeKey, error) {
	client, err := newKmsClient(keyID)
	if err != nil {
		return EnvelopeKey{}, err
	}
	generateDataKeyInput := &kms.GenerateDataKeyInput{
		KeyId: aws.String(keyID),
		EncryptionContext: aws.StringMap(map[string]string{
			"SecretName": secretID,
		}),
		NumberOfBytes: aws.Int64(32)}
	generateDataKeyOutput, err := client.GenerateDataKey(generateDataKeyInput)
	if err != nil {
		return EnvelopeKey{}, err
	}
	return EnvelopeKey{
		ResolvedID: *generateDataKeyOutput.KeyId,
		Plaintext:  generateDataKeyOutput.Plaintext,
		Ciphertext: generateDataKeyOutput.CiphertextBlob}, nil
}

// Decrypt decrypts the encrypted key.
func (k *Kms) Decrypt(keyID string, keyCiphertext []byte, secretID string) ([]byte, error) {
	client, err := newKmsClient(keyID)
	if err != nil {
		return nil, err
	}
	do, err := client.Decrypt(&kms.DecryptInput{
		EncryptionContext: aws.StringMap(map[string]string{
			"SecretName": secretID,
		}),
		CiphertextBlob: keyCiphertext,
	})
	return do.Plaintext, err
}

// Label returns kmsLabel
func (k *Kms) Label() string {
	return KmsLabel
}

func newKmsClient(arn string) (*kms.KMS, error) {
	parsed, err := NewARN(arn)
	if err != nil {
		return kms.New(session.New()), nil
	}
	return kms.New(session.New(&aws.Config{Region: &parsed.Region})), nil
}
