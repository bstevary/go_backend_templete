package auth

import "time"

type TokenGenerator interface {
	CreateToken(email string, duration time.Duration, clientIp string) (string, *Payload, error)
	ValidateToken(token string) (*Payload, error)
}
