package auth

import (
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
	ClientIP  string    `json:"client_ip"`
}

func Newpayload(Email string, duration time.Duration, clientIp string) (*Payload, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		ID:        tokenId,
		Email:     Email,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
		ClientIP:  clientIp,
	}
	return payload, nil
}

func (p *Payload) Valid() error {
	if time.Now().Before(p.IssuedAt) {
		return ErrInvalidToken
	}
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
