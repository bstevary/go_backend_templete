package db

import (
	"context"

	"github.com/bstevary/hexagonal/database/model"
)

type CreateUserTxParams struct {
	model.CreateUserParams
	AfterCreateUser func(UserID string) error
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) error {

	err := store.execTx(ctx, func(q *model.Queries) error {
		var err error

		err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreateUser(arg.ID)
	})

	return err
}
