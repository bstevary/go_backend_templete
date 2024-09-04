package db

import (
	"context"
	"fmt"

	"phcmis/databases/persist/model"

	_ "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store interface {
	model.Querier
	CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error)
	ActivateUserAccountTx(ctx context.Context, arg ActivateUserAccountTxParams) (ActivateUserAccountTxResult, error)
}
type SQLStore struct {
	connPool *pgxpool.Pool
	model.Querier
}

func NewStore(connPool *pgxpool.Pool) Store {
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
