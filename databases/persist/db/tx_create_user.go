package db

import (
	"context"

	"phcmis/databases/persist/model"
)

type CreateUserTxParams struct {
	model.CreateUserParams
	AfterCreateUser func(user model.CreateUserRow) error
}
type CreateUserTxResult struct {
	User model.CreateUserRow
}

func (store *SQLStore) CreateUserTx(ctx context.Context, arg CreateUserTxParams) (CreateUserTxResult, error) {
	var result CreateUserTxResult

	err := store.execTx(ctx, func(q *model.Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		return arg.AfterCreateUser(result.User)
	})

	return result, err
}
