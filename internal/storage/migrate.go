package storage

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed migrations/0001_init.sql
var schemaSQL string

// EnsureSchema applies the full schema. It is idempotent: every statement uses
// IF NOT EXISTS, so it is safe to run on every startup without tracking applied
// migrations.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("storage: apply schema: %w", err)
	}
	return nil
}
