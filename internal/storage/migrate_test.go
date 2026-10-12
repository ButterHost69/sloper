package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"sync"
	"testing"
)

// TestMigrateWithConcurrentMigrators runs two migrators against one fresh
// database — the state a repo container starts in, where sloper-web and sloper
// migrate together (setup/script.sh starts the API server, then the scheduler).
// Every migration must be applied once and recorded once.
func TestMigrateWithConcurrentMigrators(t *testing.T) {
	ctx := context.Background()

	// Repeat: one round can pass by luck.
	for round := 0; round < 40; round++ {
		path := filepath.Join(t.TempDir(), "sloper.sqlite")

		first, err := OpenDB(path)
		if err != nil {
			t.Fatalf("round %d: open first handle: %v", round, err)
		}
		second, err := OpenDB(path)
		if err != nil {
			_ = first.Close()
			t.Fatalf("round %d: open second handle: %v", round, err)
		}

		handles := []*sql.DB{first, second}
		errs := make([]error, len(handles))

		var wg sync.WaitGroup
		for i, db := range handles {
			wg.Add(1)
			go func(i int, db *sql.DB) {
				defer wg.Done()
				errs[i] = Migrate(ctx, db)
			}(i, db)
		}
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Fatalf("round %d: migrator %d failed: %v", round, i, err)
			}
		}

		// Every migration must be recorded exactly once, and the schema must be
		// usable afterwards.
		var applied int
		if err := first.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
			t.Fatalf("round %d: count migrations: %v", round, err)
		}
		if applied == 0 {
			t.Fatalf("round %d: no migrations recorded", round)
		}
		var duplicates int
		if err := first.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM (
				SELECT version FROM schema_migrations GROUP BY version HAVING COUNT(*) > 1
			)`).Scan(&duplicates); err != nil {
			t.Fatalf("round %d: count duplicates: %v", round, err)
		}
		if duplicates != 0 {
			t.Fatalf("round %d: %d migration versions recorded twice", round, duplicates)
		}
		if _, err := first.ExecContext(ctx, `SELECT merged_at FROM pull_requests LIMIT 1`); err != nil {
			t.Fatalf("round %d: pull_requests.merged_at missing: %v", round, err)
		}

		_ = first.Close()
		_ = second.Close()
	}
}
