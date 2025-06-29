package dbmigrate

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/null"
)

func newTestDB(t *testing.T) *sqlx.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db := sqlx.MustConnect("sqlite3", dbPath)
	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

//go:embed testdata/migrate01/*
var migrate01Dir embed.FS

func TestMigrateUp__Integration(t *testing.T) {
	db := newTestDB(t)

	MigrateUp(db, migrate01Dir, "testdata/migrate01", DatabaseSQLite3)

	assertTableExist(t, db, "auth_user")
	assertTableExist(t, db, "product")

	row, err := getMigrationRow(db)
	assert.Equal(t, nil, err)
	assert.Equal(t, null.New(SchemaMigration{
		ID:       1,
		Version:  2,
		Filename: "0002_add_product.sql",
		IsDirty:  false,
	}), row)

	// migrate up again
	MigrateUp(db, migrate01Dir, "testdata/migrate01", DatabaseSQLite3)
}

func assertTableExist(t *testing.T, db *sqlx.DB, table string) {
	t.Helper()

	var result int
	query := `SELECT COUNT(*) FROM ` + table
	err := db.Get(&result, query)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, result)
}
