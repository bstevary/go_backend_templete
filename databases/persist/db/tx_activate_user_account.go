package db

import (
	"context"

	"phcmis/databases/persist/model"

	"github.com/jackc/pgx/v5/pgtype"
)

type ActivateUserAccountTxParams struct {
	SecretCode     string
	HashedPassword string
}
type ActivateUserAccountTxResult struct {
	User                 model.User
	ActivateAccountEmail model.VarifyEmail
}

func (store *SQLStore) ActivateUserAccountTx(ctx context.Context, arg ActivateUserAccountTxParams) (ActivateUserAccountTxResult, error) {
	var result ActivateUserAccountTxResult

	err := store.execTx(ctx, func(q *model.Queries) error {
		var err error
		result.ActivateAccountEmail, err = q.UpdateActivateAccountEmail(ctx, arg.SecretCode)
		if err != nil {
			return err
		}
		result.User, err = q.AlterUserAccountStatus(ctx, model.AlterUserAccountStatusParams{
			UserID: result.ActivateAccountEmail.UserID,
			IsEmailVerified: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
			IsAccountActive: pgtype.Bool{
				Bool:  true,
				Valid: true,
			},
		})

		return err
	})

	return result, err
}
