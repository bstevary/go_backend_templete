package db

import (
	"context"
	"fmt"

	"github.com/bstevary/hexagonal/database/model"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database interface {
	model.Querier
	ResetPasswordTx(ctx context.Context, arg ResetPasswordTxParams) error
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) error
}
type SQLStore struct {
	connPool *pgxpool.Pool
	model.Querier
}

func NewDatabase(connPool *pgxpool.Pool) Database {
	return &SQLStore{
		connPool: connPool,
		Querier:  model.New(connPool),
	}
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*model.Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}
	q := model.New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err: %v rbErr: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit(ctx)
}
