package dbmigrate

import (
	"cmp"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/QuangTung97/weblib/null"
)

// SchemaMigration is the migration table stored inside client's database.
// To keep track of the last version that has migrated up to
type SchemaMigration struct {
	ID       int64  `db:"id"`
	Version  int64  `db:"version"`
	Filename string `db:"filename"`
	IsDirty  bool   `db:"is_dirty"`
}

func MigrateUp(
	db *sqlx.DB, embedDir embed.FS, migrationDir string,
	dbType DatabaseType,
) {
	entries, err := embedDir.ReadDir(migrationDir)
	if err != nil {
		panic(err)
	}

	err = doMigrateUp(
		entries,
		func() error {
			return createTableFunc(db, dbType)
		},
		func() (null.Null[SchemaMigration], error) {
			return getMigrationRow(db)
		},
		func(row SchemaMigration) error {
			return upsertRowFunc(db, dbType, row)
		},
		func(filename string) error {
			fullPath := filepath.Join(migrationDir, filename)
			data, err := embedDir.ReadFile(fullPath)
			if err != nil {
				return err
			}

			_, err = db.Exec(string(data))
			return err
		},
	)
	if err != nil {
		panic(err)
	}
}

func doMigrateUp(
	entries []fs.DirEntry,
	createTableFunc func() error,
	getRowFunc func() (null.Null[SchemaMigration], error),
	upsertRowFunc func(row SchemaMigration) error,
	runScriptFunc func(filename string) error,
) error {
	files := make([]migrateFile, 0, len(entries))
	for _, entry := range entries {
		file, err := parseMigrateFilename(entry.Name())
		if err != nil {
			return err
		}
		files = append(files, file)
	}

	if err := validateMigrateFiles(files); err != nil {
		return err
	}

	if err := createTableFunc(); err != nil {
		return err
	}

	lastMigrateRow, err := getRowFunc()
	if err != nil {
		return err
	}

	if lastMigrateRow.Valid {
		lastVersion := lastMigrateRow.Data.Version
		if lastVersion > int64(len(files)) {
			return fmt.Errorf("not found version '%04d' in migration file list", lastVersion)
		}
		files = files[lastVersion:]
	}

	for index, file := range files {
		row := SchemaMigration{
			ID:       1,
			Version:  file.version,
			Filename: file.filename,
			IsDirty:  true,
		}
		if err := upsertRowFunc(row); err != nil {
			return err
		}

		if err := runScriptFunc(file.filename); err != nil {
			return err
		}

		isLast := index >= len(files)-1
		if isLast {
			row.IsDirty = false
			if err := upsertRowFunc(row); err != nil {
				return err
			}
		}
	}

	return nil
}

type migrateFile struct {
	version  int64
	filename string
}

func parseMigrateFilename(filename string) (migrateFile, error) {
	underscoreIndex := strings.Index(filename, "_")
	if underscoreIndex <= 0 {
		return migrateFile{}, errors.New("migration filename must contain a version number and a title")
	}

	versionStr := filename[:underscoreIndex]
	if len(versionStr) < 4 {
		return migrateFile{}, fmt.Errorf("migration version number must have at least 4 characters")
	}

	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return migrateFile{}, fmt.Errorf("failed to parse migration version: '%s'", versionStr)
	}

	if version <= 0 {
		return migrateFile{}, fmt.Errorf("version number must start from 1")
	}

	return migrateFile{
		version:  version,
		filename: filename,
	}, nil
}

func validateMigrateFiles(files []migrateFile) error {
	if len(files) == 0 {
		return fmt.Errorf("migration file list must not be empty")
	}

	slices.SortFunc(files, func(a, b migrateFile) int {
		return cmp.Compare(a.version, b.version)
	})

	existedNum := map[int64]struct{}{}

	for index, file := range files {
		_, existed := existedNum[file.version]
		if existed {
			return fmt.Errorf("duplicated version number '%04d'", file.version)
		}
		existedNum[file.version] = struct{}{}

		if file.version != int64(index)+1 {
			return fmt.Errorf("missing version number '%04d'", index+1)
		}
	}

	return nil
}
