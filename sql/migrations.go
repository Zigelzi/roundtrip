package migrations

import "embed"

//go:embed schema/*.sql
var embeddedMigrations embed.FS

// GetMigrationFS returns the embedded Goose migration files so they can be
// applied without the SQL files being present on disk at runtime.
func GetMigrationFS() embed.FS {
	return embeddedMigrations
}
