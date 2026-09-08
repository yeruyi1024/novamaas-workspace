package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/QuantumNous/new-api/common"
)

const credentialEncryptionVersion = "aes-gcm-v1"

type CredentialSecret struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token,omitempty"`
}

func encryptCredential(secret CredentialSecret) (string, error) {
	key, err := storageCredentialEncryptionKey()
	if err != nil {
		return "", err
	}
	plaintext, err := common.Marshal(secret)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := aead.Seal(nil, nonce, plaintext, []byte(credentialEncryptionVersion))
	payload := append(nonce, sealed...)
	return base64.RawStdEncoding.EncodeToString(payload), nil
}

func decryptCredential(payload string) (CredentialSecret, error) {
	var secret CredentialSecret
	key, err := storageCredentialEncryptionKey()
	if err != nil {
		return secret, err
	}
	encoded, err := base64.RawStdEncoding.DecodeString(payload)
	if err != nil {
		return secret, fmt.Errorf("decode storage credential: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return secret, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return secret, err
	}
	if len(encoded) < aead.NonceSize() {
		return secret, errors.New("storage credential payload is truncated")
	}
	plaintext, err := aead.Open(nil, encoded[:aead.NonceSize()], encoded[aead.NonceSize():], []byte(credentialEncryptionVersion))
	if err != nil {
		return secret, errors.New("storage credential cannot be decrypted")
	}
	if err = common.Unmarshal(plaintext, &secret); err != nil {
		return secret, fmt.Errorf("decode storage credential data: %w", err)
	}
	return secret, nil
}

func storageCredentialEncryptionKey() ([]byte, error) {
	masterKey := strings.TrimSpace(os.Getenv("STORAGE_CREDENTIAL_ENCRYPTION_KEY"))
	if masterKey == "" {
		masterKey = strings.TrimSpace(os.Getenv("CRYPTO_SECRET"))
	}
	if masterKey == "" {
		masterKey = strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	}
	if masterKey == "" {
		return nil, errors.New("set STORAGE_CREDENTIAL_ENCRYPTION_KEY, CRYPTO_SECRET, or SESSION_SECRET before saving static storage credentials")
	}
	key := sha256.Sum256([]byte("new-api/storage-credential/v1:" + masterKey))
	return key[:], nil
}

func accessKeyHint(accessKeyID string) string {
	accessKeyID = strings.TrimSpace(accessKeyID)
	if len(accessKeyID) <= 4 {
		return "****"
	}
	return "****" + accessKeyID[len(accessKeyID)-4:]
}
