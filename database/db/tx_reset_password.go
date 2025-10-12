package db

import (
	"context"
	"time"

	"github.com/bstevary/hexagonal/database/model"

	"github.com/jackc/pgx/v5/pgtype"
)

type ResetPasswordTxParams struct {
	SecretCode     string
	HashedPassword pgtype.Text
}

func (store *SQLStore) ResetPasswordTx(ctx context.Context, arg ResetPasswordTxParams) error {

	err := store.execTx(ctx, func(q *model.Queries) error {
		var err error
		userId, err := q.UpdateVerification(ctx, arg.SecretCode)
		if err != nil {
			return err
		}
		err = q.UpdateUserSecurityInfo(ctx, model.UpdateUserSecurityInfoParams{
			ID:                 userId,
			Password:           arg.HashedPassword,
			LastPasswordChange: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			LastSecurityCheck:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			FailedAttempts:     pgtype.Int4{Int32: 0, Valid: true},
			// IsLocked:           pgtype.Bool{Bool: false, Valid: true},
		})
		if err != nil {
			return err
		}
		err = q.DeleteExpiredSessions(ctx, userId)
		if err != nil {
			return err
		}

		err = q.DeleteExpiredVerifications(ctx, userId)
		if err != nil {
			return err
		}

		return err
	})

	return err
}

func (store *SQLStore) VarifyEmailTx(ctx context.Context, arg ResetPasswordTxParams) error {

	err := store.execTx(ctx, func(q *model.Queries) error {
		var err error
		userId, err := q.UpdateVerification(ctx, arg.SecretCode)
		if err != nil {
			return err
		}
		err = q.UpdateUserSecurityInfo(ctx, model.UpdateUserSecurityInfoParams{
			ID:                 userId,
			Password:           arg.HashedPassword,
			LastPasswordChange: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			LastSecurityCheck:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			FailedAttempts:     pgtype.Int4{Int32: 0, Valid: true},
			// IsLocked:           pgtype.Bool{Bool: false, Valid: true},
		})
		if err != nil {
			return err
		}
		err = q.DeleteExpiredSessions(ctx, userId)
		if err != nil {
			return err
		}

		err = q.DeleteExpiredVerifications(ctx, userId)
		if err != nil {
			return err
		}

		return err
	})

	return err
}
