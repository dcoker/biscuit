package keymanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	myAWS "github.com/dcoker/biscuit/internal/aws"
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
func (k *Kms) GenerateEnvelopeKey(ctx context.Context, keyID string, secretID string) (EnvelopeKey, error) {
	client, err := newKmsClient(ctx, keyID)
	if err != nil {
		return EnvelopeKey{}, err
	}
	generateDataKeyInput := &kms.GenerateDataKeyInput{
		KeyId: aws.String(keyID),
		EncryptionContext: map[string]string{
			"SecretName": secretID,
		},
		NumberOfBytes: aws.Int32(32),
	}
	generateDataKeyOutput, err := client.GenerateDataKey(ctx, generateDataKeyInput)
	if err != nil {
		return EnvelopeKey{}, err
	}
	return EnvelopeKey{
		ResolvedID: *generateDataKeyOutput.KeyId,
		Plaintext:  generateDataKeyOutput.Plaintext,
		Ciphertext: generateDataKeyOutput.CiphertextBlob}, nil
}

// Decrypt decrypts the encrypted key.
func (k *Kms) Decrypt(ctx context.Context, keyID string, keyCiphertext []byte, secretID string) ([]byte, error) {
	client, err := newKmsClient(ctx, keyID)
	if err != nil {
		return nil, err
	}
	do, err := client.Decrypt(ctx, &kms.DecryptInput{
		EncryptionContext: map[string]string{
			"SecretName": secretID,
		},
		CiphertextBlob: keyCiphertext,
	})
	if err != nil {
		return []byte{}, err
	}
	return do.Plaintext, nil
}

// Label returns kmsLabel
func (k *Kms) Label() string {
	return KmsLabel
}

func newKmsClient(ctx context.Context, arn string) (*kms.Client, error) {
	cfg := myAWS.MustNewConfig(ctx)
	parsed, err := NewARN(arn)

	if err == nil {
		cfg.Region = parsed.Region
	}

	return kms.NewFromConfig(cfg), nil
}
