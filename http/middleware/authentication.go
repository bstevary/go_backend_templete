package middleware

import (
	"errors"
	"net/http"
	"strings"

	"phcmis/services/auth"
	"phcmis/services/gin_pgx_err"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(t auth.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		authorizationHeaderKey := c.GetHeader("authorization")
		if len(strings.TrimSpace(authorizationHeaderKey)) == 0 {
			err := errors.New("authorization header is not provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
			return
		}
		if len(authorizationHeaderKey) == 0 {
			err := errors.New("authorization header is not provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
			return

		}

		fields := strings.Fields(authorizationHeaderKey)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
			return
		}
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != auth.AuthorizationTypeBearer {
			err := errors.New("unsopported authorization type")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
			return
		}
		accesToken := fields[1]

		payload, err := t.ValidateToken(accesToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
			return
		}

		if payload.ClientIP != c.ClientIP() {
			err := errors.New("token bound to a different IP address")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
			return
		}

		c.Set(auth.AuthorizationPayloadkey, payload)

		c.Next()
	}
}
