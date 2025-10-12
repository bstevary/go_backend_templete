package auth

import "time"

type TokenGenerator interface {
	CreateToken(UserID string, Permissions []string, Duration time.Duration, ClientIP string, Scope string,
		ActiveReference int64, References []int64) (string, *Payload, error)
	ValidateToken(token string) (*Payload, error)
}
