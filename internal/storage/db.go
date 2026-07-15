package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const (
	DefaultDBDir  = ".sloper"
	DefaultDBName = "sloper.sqlite"

	pragmas = "?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
)

func OpenDB(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("storage: determine home dir: %w", err)
		}
		dbPath = filepath.Join(homeDir, DefaultDBDir, DefaultDBName)
	}

	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("storage: create db dir %s: %w", dir, err)
	}

	dsn := "file:" + dbPath + pragmas
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open %s: %w", dbPath, err)
	}

	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("storage: ping %s: %w", dbPath, err)
	}

	return db, nil
}
