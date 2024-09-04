// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: sessions.sql

package model

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const createSession = `-- name: CreateSession :one
INSERT INTO sessions (
  id,
  email,
  refresh_token,
  user_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING id, email, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
`

type CreateSessionParams struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, createSession,
		arg.ID,
		arg.Email,
		arg.RefreshToken,
		arg.UserAgent,
		arg.ClientIp,
		arg.IsBlocked,
		arg.ExpiresAt,
	)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const deleteSession = `-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1
`

func (q *Queries) DeleteSession(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.Exec(ctx, deleteSession, id)
	return err
}

const getSession = `-- name: GetSession :one
SELECT id, email, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at FROM sessions
WHERE id = $1
`

func (q *Queries) GetSession(ctx context.Context, id uuid.UUID) (Session, error) {
	row := q.db.QueryRow(ctx, getSession, id)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.Email,
		&i.RefreshToken,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.ExpiresAt,
		&i.CreatedAt,
	)
	return i, err
}

const updateSession = `-- name: UpdateSession :exec
UPDATE sessions
SET
  refresh_token = COALESCE($1),
  expires_at = COALESCE($2),
  is_blocked = COALESCE($3)
WHERE email = $4 
 RETURNING id, email, refresh_token, user_agent, client_ip, is_blocked, expires_at, created_at
`

type UpdateSessionParams struct {
	RefreshToken pgtype.Text        `json:"refresh_token"`
	ExpiresAt    pgtype.Timestamptz `json:"expires_at"`
	IsBlocked    pgtype.Bool        `json:"is_blocked"`
	Email        pgtype.Text        `json:"email"`
}

func (q *Queries) UpdateSession(ctx context.Context, arg UpdateSessionParams) error {
	_, err := q.db.Exec(ctx, updateSession,
		arg.RefreshToken,
		arg.ExpiresAt,
		arg.IsBlocked,
		arg.Email,
	)
	return err
}
