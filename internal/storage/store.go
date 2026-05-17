// Package storage provides SQLite persistence for scan history and collections.
package storage

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
var initSQL string

// ScanRecord represents a single scan history entry.
type ScanRecord struct {
	ID            int64
	Source        string
	ScannedAt     time.Time
	EndpointCount int
}

// CollectionRecord represents a stored collection metadata entry.
type CollectionRecord struct {
	ID        int64
	Name      string
	Source    string
	CreatedAt time.Time
	JSONPath  string
}

// Store defines the persistence interface.
type Store interface {
	SaveScan(source string, endpointCount int) error
	ListScans() ([]ScanRecord, error)
	SaveCollection(name, source, jsonPath string) error
	Close() error
}

// SQLiteStore is the SQLite-backed implementation of Store.
type SQLiteStore struct {
	db *sql.DB
}

// New opens (or creates) a SQLite database at dbPath and runs the initial migration.
func New(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if _, err := db.Exec(initSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

// SaveScan inserts a new scan record into scan_history.
func (s *SQLiteStore) SaveScan(source string, endpointCount int) error {
	_, err := s.db.Exec(
		`INSERT INTO scan_history (source, endpoint_count) VALUES (?, ?)`,
		source, endpointCount,
	)
	if err != nil {
		return fmt.Errorf("save scan: %w", err)
	}
	return nil
}

// ListScans returns all scan records ordered by most recent first.
func (s *SQLiteStore) ListScans() ([]ScanRecord, error) {
	rows, err := s.db.Query(
		`SELECT id, source, scanned_at, endpoint_count FROM scan_history ORDER BY scanned_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list scans: %w", err)
	}
	defer rows.Close()

	var records []ScanRecord
	for rows.Next() {
		var r ScanRecord
		var scannedAt string
		if err := rows.Scan(&r.ID, &r.Source, &scannedAt, &r.EndpointCount); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		r.ScannedAt, _ = time.Parse("2006-01-02 15:04:05", scannedAt)
		records = append(records, r)
	}
	return records, rows.Err()
}

// SaveCollection inserts a new collection record.
func (s *SQLiteStore) SaveCollection(name, source, jsonPath string) error {
	_, err := s.db.Exec(
		`INSERT INTO collections (name, source, json_path) VALUES (?, ?, ?)`,
		name, source, jsonPath,
	)
	if err != nil {
		return fmt.Errorf("save collection: %w", err)
	}
	return nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
