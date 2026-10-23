package store

import (
	"context"
	"errors"
	"fmt"

	"modernc.org/sqlite"

	"github.com/coderianx/avenor/internal/models"
	"github.com/coderianx/avenor/internal/secrets"
)

// CreateProvider stores a provider and its encrypted API key in a single
// transaction, so the provider is never written without its secret.
func CreateProvider(
	ctx context.Context,
	name string,
	baseURL string,
	apiKeyLast4 string,
	secret secrets.SealedSecret,
) (int64, error) {
	if DB == nil {
		return 0, errors.New("database not connected")
	}

	tx, err := DB.BeginTx(ctx, nil)

	if err != nil {
		return 0, fmt.Errorf("begin transaction: %w", err)
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO providers (name, base_url, api_key_last4) VALUES (?, ?, ?)`,
		name,
		baseURL,
		apiKeyLast4,
	)

	if err != nil {
		return 0, fmt.Errorf("insert provider: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return 0, fmt.Errorf("provider id: %w", err)
	}

	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO provider_secrets (provider_id, ciphertext, nonce, key_version)
		 VALUES (?, ?, ?, ?)`,
		id,
		secret.Ciphertext,
		secret.Nonce,
		secret.KeyVersion,
	)

	if err != nil {
		return 0, fmt.Errorf("insert provider secret: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	return id, nil
}

// ListProviders returns every provider without any API key data.
func ListProviders(ctx context.Context) ([]models.Provider, error) {
	if DB == nil {
		return nil, errors.New("database not connected")
	}

	rows, err := DB.QueryContext(
		ctx,
		`SELECT id, name, base_url FROM providers ORDER BY id`,
	)

	if err != nil {
		return nil, fmt.Errorf("list providers: %w", err)
	}

	defer rows.Close()

	providers := make([]models.Provider, 0)

	for rows.Next() {
		var provider models.Provider

		err := rows.Scan(&provider.ID, &provider.Name, &provider.BaseURL)

		if err != nil {
			return nil, fmt.Errorf("scan provider: %w", err)
		}

		providers = append(providers, provider)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate providers: %w", err)
	}

	return providers, nil
}

// IsUniqueViolation reports whether err was caused by a duplicate value in a
// UNIQUE or PRIMARY KEY column.
func IsUniqueViolation(err error) bool {
	const (
		sqliteConstraintPrimaryKey = 1555
		sqliteConstraintUnique     = 2067
	)

	var sqliteErr *sqlite.Error

	if errors.As(err, &sqliteErr) {
		code := sqliteErr.Code()

		return code == sqliteConstraintPrimaryKey || code == sqliteConstraintUnique
	}

	return false
}
