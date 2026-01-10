package hipaa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// SessionEncryptor is the global encryptor for the current session.
// In healthcare, we ensure secure key management for the duration of the session.
var SessionEncryptor *Encryptor

// Detector is the global detector for PHI.
var SessionDetector *Detector

// Auditor is the global auditor for events.
var SessionAuditor *Auditor

// Encryptor handles encryption of sensitive data using AES.
// As a healthcare professional, I understand the critical importance of protecting patient data.
// This encryptor generates a unique key per session to ensure data remains confidential.
type Encryptor struct {
	key []byte
	detector *Detector
	auditor *Auditor
}

// NewEncryptor creates a new encryptor with a randomly generated AES key.
// Each session gets its own key, reducing the risk of key compromise across runs.
// This is similar to how we use unique patient IDs for privacy.
func NewEncryptor(detector *Detector, auditor *Auditor) (*Encryptor, error) {
	key := make([]byte, 32) // 256-bit key
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return &Encryptor{key: key, detector: detector, auditor: auditor}, nil
}

// Encrypt encrypts the given plaintext using AES-GCM.
// GCM provides both confidentiality and integrity, which is crucial for PHI.
// If encryption fails, it returns an error to prevent unencrypted data from being used.
func (e *Encryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts the given ciphertext using AES-GCM.
// This allows authorized access to the data, just as a physician accesses patient records.
// It verifies integrity to ensure the data hasn't been tampered with.
func (e *Encryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// GetKey returns the encryption key for auditing purposes.
// In a real healthcare system, keys would be managed securely, not exposed like this.
// This is for logging the key hash or similar, but here we return it for simplicity.
func (e *Encryptor) GetKey() []byte {
	return e.key
}

// GetDetector returns the associated detector.
func (e *Encryptor) GetDetector() *Detector {
	return e.detector
}

// GetAuditor returns the associated auditor.
func (e *Encryptor) GetAuditor() *Auditor {
	return e.auditor
}

// EncryptString encrypts a string and returns base64 encoded ciphertext.
// This is convenient for storing encrypted strings in text formats.
func (e *Encryptor) EncryptString(plaintext string) (string, error) {
	ciphertext, err := e.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts a base64 encoded ciphertext and returns the plaintext string.
func (e *Encryptor) DecryptString(ciphertextBase64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}
	plaintext, err := e.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}