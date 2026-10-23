// Package secrets encrypts sensitive values, such as provider API keys,
// before they are written to disk.
//
// Values are encrypted with AES-256-GCM. The master key is kept in the
// operating system keyring and created on first use. On machines without a
// keyring (headless servers, CI) the base64 encoded master key can be passed
// through the AVENOR_MASTER_KEY environment variable instead.
package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	// KeyVersion identifies the master key generation used for new secrets.
	// It is stored next to every ciphertext so keys can be rotated later.
	KeyVersion = 1

	// EnvMasterKey holds an alternative, base64 encoded 32 byte master key.
	// When set it takes precedence over the OS keyring.
	EnvMasterKey = "AVENOR_MASTER_KEY"

	keyringService = "avenor"
	keyringUser    = "master-key"

	masterKeySize = 32
)

// SealedSecret is an encrypted secret together with everything needed to
// decrypt it later.
type SealedSecret struct {
	Ciphertext []byte
	Nonce      []byte
	KeyVersion int
}

// masterKey resolves and caches the master key for the process lifetime.
var masterKey = sync.OnceValues(resolveMasterKey)

// Seal encrypts plaintext with the master key.
func Seal(plaintext string) (SealedSecret, error) {
	key, err := masterKey()

	if err != nil {
		return SealedSecret{}, err
	}

	aead, err := newGCM(key)

	if err != nil {
		return SealedSecret{}, err
	}

	nonce := make([]byte, aead.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return SealedSecret{}, fmt.Errorf("secrets: generate nonce: %w", err)
	}

	return SealedSecret{
		Ciphertext: aead.Seal(nil, nonce, []byte(plaintext), nil),
		Nonce:      nonce,
		KeyVersion: KeyVersion,
	}, nil
}

// Open decrypts a sealed secret created by Seal.
func Open(sealed SealedSecret) (string, error) {
	if sealed.KeyVersion != KeyVersion {
		return "", fmt.Errorf("secrets: unsupported key version %d", sealed.KeyVersion)
	}

	key, err := masterKey()

	if err != nil {
		return "", err
	}

	aead, err := newGCM(key)

	if err != nil {
		return "", err
	}

	if len(sealed.Nonce) != aead.NonceSize() {
		return "", fmt.Errorf("secrets: invalid nonce size %d", len(sealed.Nonce))
	}

	plaintext, err := aead.Open(nil, sealed.Nonce, sealed.Ciphertext, nil)

	if err != nil {
		return "", fmt.Errorf("secrets: decrypt: %w", err)
	}

	return string(plaintext), nil
}

// resolveMasterKey returns the master key from the environment or the OS
// keyring, creating a new keyring entry when none exists yet.
func resolveMasterKey() ([]byte, error) {
	if encoded := os.Getenv(EnvMasterKey); encoded != "" {
		return decodeMasterKey(encoded, EnvMasterKey)
	}

	encoded, err := keyring.Get(keyringService, keyringUser)

	switch {
	case err == nil:
		return decodeMasterKey(encoded, "keyring")
	case errors.Is(err, keyring.ErrNotFound):
		return createMasterKey()
	default:
		return nil, fmt.Errorf(
			"secrets: read master key from keyring: %w (set %s instead)",
			err,
			EnvMasterKey,
		)
	}
}

func createMasterKey() ([]byte, error) {
	key := make([]byte, masterKeySize)

	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("secrets: generate master key: %w", err)
	}

	err := keyring.Set(
		keyringService,
		keyringUser,
		base64.StdEncoding.EncodeToString(key),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"secrets: store master key in keyring: %w (set %s instead)",
			err,
			EnvMasterKey,
		)
	}

	return key, nil
}

func decodeMasterKey(encoded string, source string) ([]byte, error) {
	key, err := base64.StdEncoding.DecodeString(encoded)

	if err != nil {
		return nil, fmt.Errorf("secrets: decode master key from %s: %w", source, err)
	}

	if len(key) != masterKeySize {
		return nil, fmt.Errorf(
			"secrets: master key from %s must be %d bytes, got %d",
			source,
			masterKeySize,
			len(key),
		)
	}

	return key, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, fmt.Errorf("secrets: new cipher: %w", err)
	}

	return cipher.NewGCM(block)
}
