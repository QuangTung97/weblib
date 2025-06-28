package dbmigrate

import (
	"embed"

	"github.com/jmoiron/sqlx"
)

// SchemaMigration is the migration table stored inside client's database.
// To keep track of the last version that has migrated up to
type SchemaMigration struct {
	ID       int64  `db:"id"`
	Version  int64  `db:"version"`
	Filename string `db:"filename"`
	IsDirty  bool   `db:"is_dirty"`
}

func MigrateUp(db *sqlx.DB, embedDir embed.FS, migrationDir string) {
}

type migrateFile struct {
	version  int64
	filename string
}

func parseMigrateFile(filename string) (migrateFile, error) {
	return migrateFile{}, nil
}

func validateMigrateFiles(files []migrateFile) error {
	return nil
}
