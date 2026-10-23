package secrets

import (
	"bytes"
	"encoding/base64"
	"strings"
	"sync"
	"testing"

	"github.com/zalando/go-keyring"
)

// resetMasterKey clears the cached master key so a test can control how it is
// resolved.
func resetMasterKey() {
	masterKey = sync.OnceValues(resolveMasterKey)
}

func TestSealOpenWithKeyring(t *testing.T) {
	keyring.MockInit()
	t.Setenv(EnvMasterKey, "")
	resetMasterKey()

	t.Cleanup(resetMasterKey)

	sealed, err := Seal("sk-test-1234567890")

	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	if bytes.Equal(sealed.Ciphertext, []byte("sk-test-1234567890")) {
		t.Fatal("ciphertext must not contain the plaintext")
	}

	if sealed.KeyVersion != KeyVersion {
		t.Fatalf("key version = %d, want %d", sealed.KeyVersion, KeyVersion)
	}

	plaintext, err := Open(sealed)

	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if plaintext != "sk-test-1234567890" {
		t.Fatalf("Open = %q, want %q", plaintext, "sk-test-1234567890")
	}
}

func TestSealOpenWithEnvKey(t *testing.T) {
	key := bytes.Repeat([]byte{0x42}, masterKeySize)

	t.Setenv(EnvMasterKey, base64.StdEncoding.EncodeToString(key))
	resetMasterKey()

	t.Cleanup(resetMasterKey)

	sealed, err := Seal("secret")

	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	// A different key must not decrypt the value.
	t.Setenv(EnvMasterKey, base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x24}, masterKeySize)))
	resetMasterKey()

	if _, err := Open(sealed); err == nil {
		t.Fatal("Open with a different master key must fail")
	}
}

func TestOpenRejectsTamperedCiphertext(t *testing.T) {
	keyring.MockInit()
	t.Setenv(EnvMasterKey, "")
	resetMasterKey()

	t.Cleanup(resetMasterKey)

	sealed, err := Seal("secret")

	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealed.Ciphertext[len(sealed.Ciphertext)-1] ^= 0xff

	if _, err := Open(sealed); err == nil {
		t.Fatal("Open must reject tampered ciphertext")
	}
}

func TestOpenRejectsUnknownKeyVersion(t *testing.T) {
	keyring.MockInit()
	t.Setenv(EnvMasterKey, "")
	resetMasterKey()

	t.Cleanup(resetMasterKey)

	sealed, err := Seal("secret")

	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	sealed.KeyVersion = KeyVersion + 1

	if _, err := Open(sealed); err == nil {
		t.Fatal("Open must reject an unknown key version")
	}
}

func TestInvalidEnvKey(t *testing.T) {
	t.Setenv(EnvMasterKey, base64.StdEncoding.EncodeToString([]byte("too-short")))
	resetMasterKey()

	t.Cleanup(resetMasterKey)

	_, err := Seal("secret")

	if err == nil {
		t.Fatal("Seal must fail with an invalid master key")
	}

	if !strings.Contains(err.Error(), EnvMasterKey) {
		t.Fatalf("error %q should mention %s", err, EnvMasterKey)
	}
}
