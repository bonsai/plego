package state

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func Open(path string) (*SQLiteStore, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create state dir: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sent (
		id TEXT PRIMARY KEY,
		sent_at INTEGER NOT NULL
	)`)
	if err != nil {
		return nil, fmt.Errorf("create table: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) IsSent(id string) bool {
	var n int
	s.db.QueryRow("SELECT COUNT(*) FROM sent WHERE id = ?", id).Scan(&n)
	return n > 0
}

func (s *SQLiteStore) MarkSent(id string) error {
	_, err := s.db.Exec(
		"INSERT OR REPLACE INTO sent (id, sent_at) VALUES (?, unixepoch())",
		id,
	)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
