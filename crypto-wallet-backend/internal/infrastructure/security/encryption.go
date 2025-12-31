package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	// AES256KeySize defines the fixed key size required for AES-256-GCM.
	AES256KeySize = 32
	// defaultScryptN controls the CPU/memory cost parameter for key derivation.
	defaultScryptN = 1 << 15
	// defaultScryptR controls the block size parameter for key derivation.
	defaultScryptR = 8
	// defaultScryptP controls the parallelisation parameter for key derivation.
	defaultScryptP = 1
	// defaultScryptKeyLen defines the derived key length used for AES-256.
	defaultScryptKeyLen = AES256KeySize
)

var (
	// ErrInvalidKeyLength indicates that the provided key does not meet AES-256 requirements.
	ErrInvalidKeyLength = errors.New("security: key must be 32 bytes for AES-256-GCM")
	// ErrInvalidCiphertext indicates malformed ciphertext payloads (e.g. missing nonce).
	ErrInvalidCiphertext = errors.New("security: invalid ciphertext payload")
)

// AESGCMEncryptor provides authenticated encryption using AES-256-GCM.
type AESGCMEncryptor struct {
	aead      cipher.AEAD
	nonceSize int
	rand      io.Reader
}

// AESGCMConfig defines the configuration for constructing an AESGCMEncryptor.
type AESGCMConfig struct {
	Key       []byte
	Random    io.Reader
}

// NewAESGCMEncryptor constructs an AESGCMEncryptor for the provided key.
func NewAESGCMEncryptor(cfg AESGCMConfig) (*AESGCMEncryptor, error) {
	if len(cfg.Key) != AES256KeySize {
		return nil, ErrInvalidKeyLength
	}

	if cfg.Random == nil {
		cfg.Random = rand.Reader
	}

	block, err := aes.NewCipher(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("security: create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("security: create gcm: %w", err)
	}

	return &AESGCMEncryptor{
		aead:      aead,
		nonceSize: aead.NonceSize(),
		rand:      cfg.Random,
	}, nil
}

// Encrypt encrypts the provided plaintext and returns nonce||ciphertext bytes.
func (e *AESGCMEncryptor) Encrypt(plaintext, additionalData []byte) ([]byte, error) {
	if e == nil || e.aead == nil {
		return nil, errors.New("security: encryptor not initialised")
	}

	nonce := make([]byte, e.nonceSize)
	if _, err := io.ReadFull(e.rand, nonce); err != nil {
		return nil, fmt.Errorf("security: read nonce: %w", err)
	}

	ciphertext := e.aead.Seal(nonce, nonce, plaintext, additionalData)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
func (e *AESGCMEncryptor) Decrypt(ciphertext, additionalData []byte) ([]byte, error) {
	if e == nil || e.aead == nil {
		return nil, errors.New("security: encryptor not initialised")
	}

	if len(ciphertext) < e.nonceSize {
		return nil, ErrInvalidCiphertext
	}

	nonce := ciphertext[:e.nonceSize]
	payload := ciphertext[e.nonceSize:]

	plaintext, err := e.aead.Open(nil, nonce, payload, additionalData)
	if err != nil {
		return nil, fmt.Errorf("security: decrypt payload: %w", err)
	}

	return plaintext, nil
}

// EncryptToString encrypts plaintext and returns a base64 encoded payload.
func (e *AESGCMEncryptor) EncryptToString(plaintext, additionalData []byte) (string, error) {
	bytes, err := e.Encrypt(plaintext, additionalData)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// DecryptString decodes a base64 payload and decrypts the contents.
func (e *AESGCMEncryptor) DecryptString(payload string, additionalData []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("security: decode payload: %w", err)
	}
	return e.Decrypt(data, additionalData)
}

// GenerateRandomKey returns a cryptographically secure random key suitable for AES-256-GCM.
func GenerateRandomKey() ([]byte, error) {
	key := make([]byte, AES256KeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("security: generate key: %w", err)
	}
	return key, nil
}

// DeriveKeyFromPassphrase derives a key using scrypt from the provided passphrase and salt.
func DeriveKeyFromPassphrase(passphrase, salt []byte) ([]byte, error) {
	if len(passphrase) == 0 {
		return nil, errors.New("security: passphrase is required")
	}
	if len(salt) == 0 {
		return nil, errors.New("security: salt is required")
	}

	key, err := scrypt.Key(passphrase, salt, defaultScryptN, defaultScryptR, defaultScryptP, defaultScryptKeyLen)
	if err != nil {
		return nil, fmt.Errorf("security: derive key: %w", err)
	}
	return key, nil
}
