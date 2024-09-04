package auth

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

type pasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// CreateToken implements TokenGenerator.
func (maker *pasetoMaker) CreateToken(email string, duration time.Duration, clientIp string) (string, *Payload, error) {
	payload, err := Newpayload(email, duration, clientIp)
	if err != nil {
		return "", payload, fmt.Errorf("failed to create payload %w", err)
	}
	token, err := maker.paseto.Encrypt(maker.symmetricKey, payload, nil)
	return token, payload, err
}

// ValidateToken implements TokenGenerator.
func (maker *pasetoMaker) ValidateToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := maker.paseto.Decrypt(token, maker.symmetricKey, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload %w", err)
	}
	err = payload.Valid()
	if err != nil {
		return nil, err
	}
	return payload, nil
}

func NewPasetoMaker(symmetricKey string) (TokenGenerator, error) {
	sm := []byte(symmetricKey)
	if len(sm) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size")
	}
	maker := &pasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: sm,
	}
	return maker, nil
}
