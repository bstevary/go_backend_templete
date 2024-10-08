// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type Session struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	RefreshToken string    `json:"refresh_token"`
	UserAgent    string    `json:"user_agent"`
	ClientIp     string    `json:"client_ip"`
	IsBlocked    bool      `json:"is_blocked"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type User struct {
	UserID            int64              `json:"user_id"`
	FirstName         string             `json:"first_name"`
	LastName          string             `json:"last_name"`
	Email             string             `json:"email"`
	Gender            string             `json:"gender"`
	HashedPassword    string             `json:"hashed_password"`
	PasswordChangedAt pgtype.Timestamptz `json:"password_changed_at"`
	DateOfBirth       time.Time          `json:"date_of_birth"`
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         pgtype.Timestamptz `json:"updated_at"`
	IsEmailVerified   bool               `json:"is_email_verified"`
	IsAccountActive   bool               `json:"is_account_active"`
}

type VarifyEmail struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	Email      string    `json:"email"`
	SecretCode string    `json:"secret_code"`
	IsUsed     bool      `json:"is_used"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiredAt  time.Time `json:"expired_at"`
}
