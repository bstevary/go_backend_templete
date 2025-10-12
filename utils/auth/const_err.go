package auth

import "errors"

// const (
// 	AuthorizationHeaderKey  = "authorization"
// 	AuthorizationTypeBearer = "bearer"
// 	AuthorizationPayloadkey = "authorization_payload"
// )

const (
	AuthHeaderKey   = "authorization"
	BearerAuthToken = "bearer"
	AuthKey         = "auth"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)
