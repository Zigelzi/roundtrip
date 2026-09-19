package db

import (
	"database/sql"
	"errors"
	"fmt"

	migrations "github.com/Zigelzi/roundtrip/sql"
	"github.com/pressly/goose/v3"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// InitDB opens (creating if needed) the SQLite database at dbPath and verifies
// the connection. The pure-Go modernc.org/sqlite driver is used so the binary
// cross-compiles to the Pi without a C toolchain.
//
// The DSN pragmas are applied to every pooled connection: busy_timeout makes a
// concurrent writer wait rather than fail with SQLITE_BUSY (Roundtrip is a
// two-writer app), WAL improves read/write concurrency, and foreign_keys(on)
// enforces the constraints the real domain model will rely on (SQLite defaults
// it OFF).
func InitDB(dbPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(on)", dbPath)
	database, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}
	if err := database.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return database, nil
}

// RunMigrations applies the embedded Goose migrations to the database.
func RunMigrations(database *sql.DB) error {
	goose.SetBaseFS(migrations.GetMigrationFS())
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}
	if err := goose.Up(database, "schema"); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// IsUniqueViolation reports whether err is SQLite rejecting a row that breaks
// a UNIQUE constraint.
func IsUniqueViolation(err error) bool {
	var sqliteErr *sqlite.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}
