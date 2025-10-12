package auth

import (
	"time"

	"github.com/google/uuid"
)

type Payload struct {
	ID              uuid.UUID `json:"id"`
	UserID          string    `json:"user_id"`
	Permissions     []string  `json:"permissions"`
	IssuedAt        time.Time `json:"issued_at"`
	ExpiredAt       time.Time `json:"expired_at"`
	ClientIP        string    `json:"client_ip"`
	References      []int64   `json:"references"`
	Scope           string    `json:"scope"`
	ActiveReference int64     `json:"active_reference"`
}

func Newpayload(UserID string, Permissions []string, Duration time.Duration, ClientIP string, Scope string,
	ActiveReference int64, References []int64) (*Payload, error) {
	TokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		ID:              TokenID,
		UserID:          UserID,
		Permissions:     Permissions,
		IssuedAt:        time.Now(),
		ExpiredAt:       time.Now().Add(Duration),
		ClientIP:        ClientIP,
		Scope:           Scope,
		References:      References,
		ActiveReference: ActiveReference,
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
