package dblib

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/jmoiron/sqlx"

	"github.com/QuangTung97/weblib/null"
)

type Provider interface {
	Transact(ctx context.Context, fn func(ctx context.Context) error) error
	Readonly(ctx context.Context) context.Context
	Autocommit(ctx context.Context) context.Context
}

type Readonly interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	Rebind(query string) string
}

type Transaction interface {
	Readonly

	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error)
	NamedQuery(query string, arg any) (*sqlx.Rows, error)
}

var _ Transaction = &sqlx.DB{}
var _ Transaction = &sqlx.Tx{}

func GetReadonly(ctx context.Context) Readonly {
	val, ok := getFromContext(ctx)
	if !ok {
		panic("Missing call to method of dblib.Provider")
	}
	return val.tx
}

func GetTx(ctx context.Context) Transaction {
	val, ok := getFromContext(ctx)
	if !ok {
		panic("Missing call to method of dblib.Provider")
	}

	if val.isReadonly {
		panic("Can not get transaction object from context of Provider.Readonly")
	}

	return val.tx
}

func NullGet[T any](ctx context.Context, query string, args ...any) (null.Null[T], error) {
	var result T
	err := GetReadonly(ctx).GetContext(ctx, &result, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return null.Null[T]{}, nil
		}
		return null.Null[T]{}, err
	}
	return null.New(result), nil
}

func NewProvider(db *sqlx.DB) Provider {
	return &providerImpl{
		db: db,
	}
}

type providerImpl struct {
	db *sqlx.DB
}

func (p *providerImpl) Transact(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	val, ok := getFromContext(ctx)
	if ok && !val.isReadonly {
		return fn(ctx)
	}

	tx, err := p.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			err = fmt.Errorf("panic: %v", r)
		}
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	val = &contextValueType{
		isReadonly: false,
		tx:         tx,
	}
	ctx = setToContext(ctx, val)

	err = fn(ctx)
	return err
}

func (p *providerImpl) Readonly(ctx context.Context) context.Context {
	return setToContext(ctx, &contextValueType{
		isReadonly: true,
		tx:         p.db,
	})
}

func (p *providerImpl) Autocommit(ctx context.Context) context.Context {
	return setToContext(ctx, &contextValueType{
		isReadonly: false,
		tx:         p.db,
	})
}
