package dbmigrate

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/QuangTung97/weblib/null"
)

type DatabaseType int

const (
	DatabaseSQLite3 DatabaseType = iota + 1
	DatabaseMySQL
)

const SQLiteCreateTableQuery = `
CREATE TABLE IF NOT EXISTS schema_migration (
    id INTEGER NOT NULL PRIMARY KEY,
    version INTEGER NOT NULL,
    filename TEXT NOT NULL,
	is_dirty INTEGER NOT NULL
) STRICT;
`

const MySQLCreateTableQuery = `
CREATE TABLE IF NOT EXISTS schema_migration (
	id INTEGER NOT NULL PRIMARY KEY,
    version INTEGER NOT NULL,
    filename VARCHAR(1024) NOT NULL,
	is_dirty BOOLEAN NOT NULL
);
`

func createTableFunc(db *sqlx.DB, dbType DatabaseType) error {
	var query string
	switch dbType {
	case DatabaseSQLite3:
		query = SQLiteCreateTableQuery
	case DatabaseMySQL:
		query = MySQLCreateTableQuery
	default:
		return fmt.Errorf("unsupported database type: %v", dbType)
	}

	_, err := db.Exec(query)
	return err
}

const SQLite3UpsertRowQuery = `
INSERT INTO schema_migration (
    id, version, filename, is_dirty
) VALUES (
    :id, :version, :filename, :is_dirty
)
ON CONFLICT (id) DO UPDATE SET
	version = EXCLUDED.version,
	filename = EXCLUDED.filename,
	is_dirty = EXCLUDED.is_dirty
`

const MySQLUpsertRowQuery = `
INSERT INTO schema_migration (
    id, version, filename, is_dirty
) VALUES (
    :id, :version, :filename, :is_dirty
) AS new
ON DUPLICATED KEY UPDATE
    version = new.version,
	filename = new.filename,
	is_dirty = new.is_dirty
`

func upsertRowFunc(db *sqlx.DB, dbType DatabaseType, row SchemaMigration) error {
	var query string

	switch dbType {
	case DatabaseSQLite3:
		query = SQLite3UpsertRowQuery
	case DatabaseMySQL:
		query = MySQLUpsertRowQuery
	default:
		return fmt.Errorf("unsupported database type: %v", dbType)
	}

	_, err := db.NamedExec(query, row)
	return err
}

func getMigrationRow(db *sqlx.DB) (null.Null[SchemaMigration], error) {
	query := `
SELECT id, version, filename, is_dirty FROM schema_migration
WHERE id = 1
`
	var result []SchemaMigration
	if err := db.Select(&result, query); err != nil {
		return null.Null[SchemaMigration]{}, err
	}
	if len(result) == 0 {
		return null.Null[SchemaMigration]{}, nil
	}
	return null.New(result[0]), nil
}
