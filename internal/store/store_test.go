package store

import (
	"bytes"
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/coderianx/avenor/internal/secrets"
)

func openTestDB(t *testing.T) {
	t.Helper()

	if err := ConnectDB(filepath.Join(t.TempDir(), "avenor.db")); err != nil {
		t.Fatalf("ConnectDB: %v", err)
	}

	t.Cleanup(func() {
		if err := CloseDB(); err != nil {
			t.Errorf("CloseDB: %v", err)
		}
	})

	if err := CreateTables(); err != nil {
		t.Fatalf("CreateTables: %v", err)
	}
}

func TestConnectDBCreatesPrivateFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "avenor.db")

	if err := ConnectDB(path); err != nil {
		t.Fatalf("ConnectDB: %v", err)
	}

	t.Cleanup(func() {
		if err := CloseDB(); err != nil {
			t.Errorf("CloseDB: %v", err)
		}
	})

	info, err := os.Stat(path)

	if err != nil {
		t.Fatalf("stat database file: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("database permissions = %o, want 600", perm)
	}
}

func TestListProviders(t *testing.T) {
	t.Setenv(
		secrets.EnvMasterKey,
		base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x13}, 32)),
	)

	openTestDB(t)

	ctx := context.Background()

	empty, err := ListProviders(ctx)

	if err != nil {
		t.Fatalf("ListProviders on empty database: %v", err)
	}

	if empty == nil {
		t.Fatal("ListProviders must return an empty slice, not nil")
	}

	if len(empty) != 0 {
		t.Fatalf("providers = %d, want 0", len(empty))
	}

	sealed, err := secrets.Seal("sk-test")

	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	firstID, err := CreateProvider(ctx, "first", "https://first.test/v1", "t", sealed)

	if err != nil {
		t.Fatalf("CreateProvider first: %v", err)
	}

	secondID, err := CreateProvider(ctx, "second", "https://second.test/v1", "", sealed)

	if err != nil {
		t.Fatalf("CreateProvider second: %v", err)
	}

	providers, err := ListProviders(ctx)

	if err != nil {
		t.Fatalf("ListProviders: %v", err)
	}

	if len(providers) != 2 {
		t.Fatalf("providers = %d, want 2", len(providers))
	}

	want := []struct {
		id      int64
		name    string
		baseURL string
	}{
		{firstID, "first", "https://first.test/v1"},
		{secondID, "second", "https://second.test/v1"},
	}

	for i, expected := range want {
		got := providers[i]

		if got.ID != expected.id || got.Name != expected.name || got.BaseURL != expected.baseURL {
			t.Fatalf("provider[%d] = %+v, want %+v", i, got, expected)
		}
	}
}

func TestCreateProviderEncryptsAPIKey(t *testing.T) {
	const apiKey = "sk-test-abcdef1234567890"

	t.Setenv(
		secrets.EnvMasterKey,
		base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x13}, 32)),
	)

	openTestDB(t)

	sealed, err := secrets.Seal(apiKey)

	if err != nil {
		t.Fatalf("Seal: %v", err)
	}

	ctx := context.Background()

	id, err := CreateProvider(ctx, "test", "https://api.test/v1", "7890", sealed)

	if err != nil {
		t.Fatalf("CreateProvider: %v", err)
	}

	var (
		ciphertext []byte
		nonce      []byte
		keyVersion int
	)

	err = DB.QueryRowContext(
		ctx,
		`SELECT ciphertext, nonce, key_version FROM provider_secrets WHERE provider_id = ?`,
		id,
	).Scan(&ciphertext, &nonce, &keyVersion)

	if err != nil {
		t.Fatalf("select provider secret: %v", err)
	}

	if bytes.Contains(ciphertext, []byte(apiKey)) {
		t.Fatal("stored ciphertext contains the plaintext api key")
	}

	plaintext, err := secrets.Open(secrets.SealedSecret{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		KeyVersion: keyVersion,
	})

	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	if plaintext != apiKey {
		t.Fatalf("Open = %q, want %q", plaintext, apiKey)
	}

	// A duplicate base_url must be reported and rolled back as a whole.
	_, err = CreateProvider(ctx, "other", "https://api.test/v1", "", sealed)

	if !IsUniqueViolation(err) {
		t.Fatalf("duplicate base_url error = %v, want unique violation", err)
	}

	var secretsCount int

	if err := DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM provider_secrets`).Scan(&secretsCount); err != nil {
		t.Fatalf("count provider secrets: %v", err)
	}

	if secretsCount != 1 {
		t.Fatalf("provider_secrets rows = %d, want 1", secretsCount)
	}
}
