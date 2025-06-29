package dbmigrate

import (
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/null"
	"github.com/QuangTung97/weblib/sliceutil"
)

func TestParseMigrateFileName(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		output, err := parseMigrateFilename("0012_init.sql")
		assert.Equal(t, nil, err)
		assert.Equal(t, migrateFile{
			version:  12,
			filename: "0012_init.sql",
		}, output)
	})

	t.Run("missing number", func(t *testing.T) {
		output, err := parseMigrateFilename("init.sql")
		assert.Equal(t, errors.New("migration filename must contain a version number and a title"), err)
		assert.Equal(t, migrateFile{}, output)
	})

	t.Run("version number less than 4 chars", func(t *testing.T) {
		output, err := parseMigrateFilename("002_init.sql")
		assert.Equal(t, errors.New("migration version number must have at least 4 characters"), err)
		assert.Equal(t, migrateFile{}, output)
	})

	t.Run("failed to parse version", func(t *testing.T) {
		output, err := parseMigrateFilename("00xx_init.sql")
		assert.Equal(t, errors.New("failed to parse migration version: '00xx'"), err)
		assert.Equal(t, migrateFile{}, output)
	})

	t.Run("version is zero", func(t *testing.T) {
		output, err := parseMigrateFilename("0000_init.sql")
		assert.Equal(t, errors.New("version number must start from 1"), err)
		assert.Equal(t, migrateFile{}, output)
	})
}

func TestValidateMigrateFiles(t *testing.T) {
	t.Run("single", func(t *testing.T) {
		err := validateMigrateFiles([]migrateFile{
			{
				version:  1,
				filename: "0001_init.sql",
			},
		})
		assert.Equal(t, nil, err)
	})

	t.Run("single, not start at 1", func(t *testing.T) {
		err := validateMigrateFiles([]migrateFile{
			{
				version:  2,
				filename: "0002_init.sql",
			},
		})
		assert.Equal(t, errors.New("missing version number '0001'"), err)
	})

	t.Run("version number not consecutive", func(t *testing.T) {
		err := validateMigrateFiles([]migrateFile{
			{
				version:  4,
				filename: "0004_add_other.sql",
			},
			{
				version:  1,
				filename: "0001_init.sql",
			},
			{
				version:  2,
				filename: "0002_add_user.sql",
			},
		})
		assert.Equal(t, errors.New("missing version number '0003'"), err)
	})

	t.Run("duplicated version number", func(t *testing.T) {
		err := validateMigrateFiles([]migrateFile{
			{
				version:  1,
				filename: "0001_init.sql",
			},
			{
				version:  2,
				filename: "0002_add_user.sql",
			},
			{
				version:  2,
				filename: "0002_add_other.sql",
			},
		})
		assert.Equal(t, errors.New("duplicated version number '0002'"), err)
	})

	t.Run("empty", func(t *testing.T) {
		err := validateMigrateFiles(nil)
		assert.Equal(t, errors.New("migration file list must not be empty"), err)
	})
}

type migrateTest struct {
	createErr   error
	createCalls int

	getRowValue null.Null[SchemaMigration]
	getRowErr   error
	getRowCalls int

	upsertErr    error
	upsertInputs []SchemaMigration

	runScriptErr    error
	runScriptInputs []string

	actions []string
}

func newMigrateTest() *migrateTest {
	m := &migrateTest{}

	return m
}

func (m *migrateTest) addAction(action string) {
	m.actions = append(m.actions, action)
}

type dirEntryTest struct {
	fs.DirEntry
	name string
}

func (e dirEntryTest) Name() string {
	return e.name
}

func (m *migrateTest) executeMigrate(entries ...dirEntryTest) error {
	inputEntries := sliceutil.Map(entries, func(e dirEntryTest) fs.DirEntry {
		return e
	})
	return doMigrateUp(
		inputEntries,
		func() error {
			m.createCalls++
			return m.createErr
		},
		func() (null.Null[SchemaMigration], error) {
			m.getRowCalls++
			return m.getRowValue, m.getRowErr
		},
		func(row SchemaMigration) error {
			m.addAction("upsert_row")
			m.upsertInputs = append(m.upsertInputs, row)
			return m.upsertErr
		},
		func(filename string) error {
			m.addAction("run_script")
			m.runScriptInputs = append(m.runScriptInputs, filename)
			return m.runScriptErr
		},
	)
}

func TestDoMigrateUp(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		m := newMigrateTest()

		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
		)
		assert.Equal(t, nil, err)

		// check calls
		assert.Equal(t, 1, m.createCalls)
		assert.Equal(t, 1, m.getRowCalls)
		assert.Equal(t, []SchemaMigration{
			{
				ID:       1,
				Version:  1,
				Filename: "0001_init.sql",
				IsDirty:  true,
			},
			{
				ID:       1,
				Version:  1,
				Filename: "0001_init.sql",
				IsDirty:  false,
			},
		}, m.upsertInputs)

		// check run scripts
		assert.Equal(t, []string{
			"0001_init.sql",
		}, m.runScriptInputs)

		// check actions
		assert.Equal(t, []string{
			"upsert_row",
			"run_script",
			"upsert_row",
		}, m.actions)
	})

	t.Run("empty dir", func(t *testing.T) {
		m := newMigrateTest()

		err := m.executeMigrate()
		assert.Equal(t, errors.New("migration file list must not be empty"), err)
	})

	t.Run("invalid file name format", func(t *testing.T) {
		m := newMigrateTest()

		err := m.executeMigrate(dirEntryTest{
			name: "invalid.sql",
		})
		assert.Equal(t, errors.New("migration filename must contain a version number and a title"), err)
	})

	t.Run("create table error", func(t *testing.T) {
		m := newMigrateTest()

		m.createErr = errors.New("create table error")

		err := m.executeMigrate(dirEntryTest{
			name: "0001_init.sql",
		})
		assert.Equal(t, errors.New("create table error"), err)

		assert.Equal(t, 0, m.getRowCalls)
		assert.Equal(t, []string(nil), m.actions)
	})

	t.Run("multiple migration files", func(t *testing.T) {
		m := newMigrateTest()

		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
		)
		assert.Equal(t, nil, err)

		// check calls
		assert.Equal(t, 1, m.createCalls)
		assert.Equal(t, 1, m.getRowCalls)
		assert.Equal(t, []SchemaMigration{
			{
				ID:       1,
				Version:  1,
				Filename: "0001_init.sql",
				IsDirty:  true,
			},
			{
				ID:       1,
				Version:  2,
				Filename: "0002_add_user.sql",
				IsDirty:  true,
			},
			{
				ID:       1,
				Version:  2,
				Filename: "0002_add_user.sql",
				IsDirty:  false,
			},
		}, m.upsertInputs)

		// check run scripts
		assert.Equal(t, []string{
			"0001_init.sql",
			"0002_add_user.sql",
		}, m.runScriptInputs)

		// check actions
		assert.Equal(t, []string{
			"upsert_row",
			"run_script",
			"upsert_row",
			"run_script",
			"upsert_row",
		}, m.actions)
	})

	t.Run("multiple migration files, with previous position", func(t *testing.T) {
		m := newMigrateTest()

		m.getRowValue = null.New(SchemaMigration{
			ID:       1,
			Version:  2,
			Filename: "0002_add_user.sql",
			IsDirty:  false,
		})

		// do migrate up
		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
			dirEntryTest{name: "0003_add_index.sql"},
			dirEntryTest{name: "0004_add_permission.sql"},
		)
		assert.Equal(t, nil, err)

		// check calls
		assert.Equal(t, []SchemaMigration{
			{
				ID:       1,
				Version:  3,
				Filename: "0003_add_index.sql",
				IsDirty:  true,
			},
			{
				ID:       1,
				Version:  4,
				Filename: "0004_add_permission.sql",
				IsDirty:  true,
			},
			{
				ID:       1,
				Version:  4,
				Filename: "0004_add_permission.sql",
				IsDirty:  false,
			},
		}, m.upsertInputs)

		// check run scripts
		assert.Equal(t, []string{
			"0003_add_index.sql",
			"0004_add_permission.sql",
		}, m.runScriptInputs)
	})

	t.Run("multiple migration files, prev version bigger than input filenames", func(t *testing.T) {
		m := newMigrateTest()

		m.getRowValue = null.New(SchemaMigration{
			ID:       1,
			Version:  4,
			Filename: "0004_add_perm.sql",
			IsDirty:  false,
		})

		// do migrate up
		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
			dirEntryTest{name: "0003_add_index.sql"},
		)
		assert.Equal(t, fmt.Errorf("not found version '0004' in migration file list"), err)

		// check calls
		assert.Equal(t, []SchemaMigration(nil), m.upsertInputs)
	})

	t.Run("multiple migration files, with previous position, no action", func(t *testing.T) {
		m := newMigrateTest()

		m.getRowValue = null.New(SchemaMigration{
			ID:       1,
			Version:  4,
			Filename: "0004_add_permission.sql",
			IsDirty:  false,
		})

		// do migrate up
		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
			dirEntryTest{name: "0003_add_index.sql"},
			dirEntryTest{name: "0004_add_permission.sql"},
		)
		assert.Equal(t, nil, err)

		// check calls
		assert.Equal(t, []SchemaMigration(nil), m.upsertInputs)
	})

	t.Run("get row error", func(t *testing.T) {
		m := newMigrateTest()

		m.getRowErr = errors.New("get row error")

		// do migrate up
		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
		)
		assert.Equal(t, m.getRowErr, err)

		// check calls
		assert.Equal(t, []SchemaMigration(nil), m.upsertInputs)
	})

	t.Run("upsert error", func(t *testing.T) {
		m := newMigrateTest()

		m.upsertErr = errors.New("upsert error")

		// do migrate up
		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
		)
		assert.Equal(t, m.upsertErr, err)

		// check calls
		assert.Equal(t, []SchemaMigration{
			{
				ID:       1,
				Version:  1,
				Filename: "0001_init.sql",
				IsDirty:  true,
			},
		}, m.upsertInputs)

		assert.Equal(t, []string{
			"upsert_row",
		}, m.actions)
	})

	t.Run("run script error", func(t *testing.T) {
		m := newMigrateTest()

		m.runScriptErr = errors.New("run script error")

		// do migrate up
		err := m.executeMigrate(
			dirEntryTest{name: "0001_init.sql"},
			dirEntryTest{name: "0002_add_user.sql"},
		)
		assert.Equal(t, m.runScriptErr, err)

		// check calls
		assert.Equal(t, []SchemaMigration{
			{
				ID:       1,
				Version:  1,
				Filename: "0001_init.sql",
				IsDirty:  true,
			},
		}, m.upsertInputs)

		assert.Equal(t, []string{
			"upsert_row",
			"run_script",
		}, m.actions)
	})
}
