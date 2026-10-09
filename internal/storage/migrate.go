package storage

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/ButterHost69/sloper/internal/logger"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func Migrate(ctx context.Context, db *sql.DB) error {
	log := logger.Default()

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
		)
	`); err != nil {
		return fmt.Errorf("storage: create schema_migrations table: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("storage: read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")

		// Cheap pre-check: an already-migrated database should not open a write
		// transaction per migration. It is not the guard against concurrency —
		// the claim below is.
		var exists bool
		err := db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("storage: check migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("storage: read migration %s: %w", file, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("storage: begin tx for %s: %w", version, err)
		}

		// Claim the version before applying it. Two sloper processes share one
		// database inside a repo container (setup/script.sh starts sloper-web,
		// then sloper), and both migrate: the INSERT takes SQLite's write lock,
		// so a competing migrator waits here and then finds the row present
		// instead of re-running an ALTER TABLE that is already applied.
		claimed, err := claimMigration(ctx, tx, version)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("storage: record migration %s: %w", version, err)
		}
		if !claimed {
			_ = tx.Rollback()
			log.Debug("storage: migration applied by another process", logger.WithStage(version))
			continue
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("storage: exec migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("storage: commit migration %s: %w", version, err)
		}

		log.Info("storage: applied migration", logger.WithStage(version))
	}

	return nil
}

// claimMigration records the version inside tx and reports whether this process
// won the claim. The claim and the migration body commit together, so a failed
// migration leaves no claim behind and is retried on the next start.
func claimMigration(ctx context.Context, tx *sql.Tx, version string) (bool, error) {
	res, err := tx.ExecContext(ctx,
		"INSERT OR IGNORE INTO schema_migrations (version) VALUES (?)", version)
	if err != nil {
		return false, err
	}
	claimed, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return claimed > 0, nil
}
