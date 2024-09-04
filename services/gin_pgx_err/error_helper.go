package gin_pgx_err

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type FieldError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "Field is required."
	case "email":
		return "Invalid email format."
	case "alphanum":
		return "Field must be alphanumeric."
	case "min":
		return "Minimum length not met."
	case "max":
		return " Maximum length exceeded."

	}

	return " Field is invalid."
}

func ErrorResponse(err error) gin.H {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]FieldError, len(ve))
		for i, fe := range ve {
			// Convert field name from CamelCase to snake_case
			snakeFieldName := strcase.ToSnake(fe.Field())
			// Prepend the field name to the message
			message := msgForTag(fe.Tag())
			out[i] = FieldError{Field: snakeFieldName, Msg: message}
		}
		return gin.H{"error": out}
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Handle PostgreSQL errors specifically
		return handleDatabaseError(pgErr)
	}
	return gin.H{"error": []FieldError{{Field: "root", Msg: err.Error()}}}
}
