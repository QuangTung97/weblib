package dblib

import (
	"context"
	"embed"
	"errors"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"github.com/QuangTung97/weblib/dbmigrate"
)

//go:embed testdata/migrate/*
var migrateDir embed.FS

func newTestDB(t *testing.T) *sqlx.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db := sqlx.MustConnect("sqlite3", dbPath)
	t.Cleanup(func() {
		_ = db.Close()
	})

	dbmigrate.MigrateUp(db, migrateDir, "testdata/migrate", dbmigrate.DatabaseSQLite3)

	return db
}

type authUser struct {
	ID        int64  `db:"id"`
	Username  string `db:"username"`
	CreatedAt int64  `db:"created_at"`
}

func insertAuthUser(ctx context.Context, user *authUser) {
	query := `
INSERT INTO auth_user (username, created_at)
VALUES (:username, :created_at)
`
	result, err := GetTx(ctx).NamedExecContext(ctx, query, user)
	if err != nil {
		panic(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		panic(err)
	}
	user.ID = id
}

func getAuthUser(ctx context.Context, id int64) authUser {
	query := `SELECT id, username, created_at FROM auth_user WHERE id = ?`
	result, err := NullGet[authUser](ctx, query, id)
	if err != nil {
		panic(err)
	}
	return result.Data
}

func TestProvider__Transact__Insert_Then_Get(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}

	err := provider.Transact(context.Background(), func(ctx context.Context) error {
		insertAuthUser(ctx, &user01)
		assert.Equal(t, authUser{
			ID:        1,
			Username:  "user01",
			CreatedAt: 2001,
		}, getAuthUser(ctx, user01.ID))
		return nil
	})
	assert.Equal(t, nil, err)
}

func TestProvider__Transact__Then_Readonly(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}
	user02 := authUser{
		Username:  "user02",
		CreatedAt: 2002,
	}

	// insert
	err := provider.Transact(context.Background(), func(ctx context.Context) error {
		insertAuthUser(ctx, &user01)
		insertAuthUser(ctx, &user02)
		return nil
	})
	assert.Equal(t, nil, err)

	// get
	readCtx := provider.Readonly(context.Background())
	assert.Equal(t, user01, getAuthUser(readCtx, user01.ID))

	assert.PanicsWithValue(t, "Can not get transaction object from context of Provider.Readonly", func() {
		user03 := authUser{
			Username:  "user03",
			CreatedAt: 2003,
		}
		insertAuthUser(readCtx, &user03)
	})
}

func TestProvider__Autocommit(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)
	ctx := provider.Autocommit(context.Background())

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}

	insertAuthUser(ctx, &user01)
	assert.Equal(t, user01, getAuthUser(ctx, user01.ID))
}

func TestProvider__EmptyContext(t *testing.T) {
	assert.PanicsWithValue(t, "Missing call to method of dblib.Provider", func() {
		getAuthUser(context.Background(), 1)
	})
}

func TestProvider__Transact__With_Error(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}

	// insert
	err := provider.Transact(context.Background(), func(ctx context.Context) error {
		insertAuthUser(ctx, &user01)
		return errors.New("handle error")
	})
	assert.Equal(t, errors.New("handle error"), err)

	// get
	readCtx := provider.Readonly(context.Background())
	assert.Equal(t, authUser{}, getAuthUser(readCtx, user01.ID))
}

func TestProvider__Transact__Panic_Inside(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}

	// insert
	err := provider.Transact(context.Background(), func(ctx context.Context) error {
		insertAuthUser(ctx, &user01)
		panic("some value")
	})
	assert.Equal(t, errors.New("panic: some value"), err)

	// get
	readCtx := provider.Readonly(context.Background())
	assert.Equal(t, authUser{}, getAuthUser(readCtx, user01.ID))
}

func TestProvider__Transact_Inside_Transact__Success(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}
	user02 := authUser{
		Username:  "user02",
		CreatedAt: 2002,
	}

	// insert
	err := provider.Transact(context.Background(), func(ctx context.Context) error {
		insertAuthUser(ctx, &user01)

		return provider.Transact(ctx, func(ctx context.Context) error {
			insertAuthUser(ctx, &user02)
			return nil
		})
	})
	assert.Equal(t, nil, err)

	// get
	readCtx := provider.Readonly(context.Background())
	assert.Equal(t, user01, getAuthUser(readCtx, user01.ID))
	assert.Equal(t, user02, getAuthUser(readCtx, user02.ID))
}

func TestProvider__Transact_Inside_Transact__Rollback(t *testing.T) {
	db := newTestDB(t)
	provider := NewProvider(db)

	user01 := authUser{
		Username:  "user01",
		CreatedAt: 2001,
	}
	user02 := authUser{
		Username:  "user02",
		CreatedAt: 2002,
	}

	// insert
	err := provider.Transact(context.Background(), func(ctx context.Context) error {
		insertAuthUser(ctx, &user01)

		err := provider.Transact(ctx, func(ctx context.Context) error {
			insertAuthUser(ctx, &user02)
			return nil
		})
		assert.Equal(t, nil, err)

		return errors.New("test rollback")
	})
	assert.Equal(t, errors.New("test rollback"), err)

	// get
	readCtx := provider.Readonly(context.Background())
	assert.Equal(t, authUser{}, getAuthUser(readCtx, user01.ID))
	assert.Equal(t, authUser{}, getAuthUser(readCtx, user02.ID))
}
