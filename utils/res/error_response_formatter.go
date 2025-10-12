package res

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rs/zerolog/log"
)

type FieldError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
}

// UnifiedErrorResponse represents the desired output structure
type UnifiedErrorResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors"`
}

func msgForTag(tag string) string {
	switch tag {
	case "required":
		return "Field is required."
	case "email":
		return "Invalid email format."
	case "alphanum":
		return "Field must be alphanumeric."
	case "alpha":
		return "Field must be alphabetic."
	case "min":
		return "Minimum length not met."
	case "max":
		return "Maximum length exceeded."
	default:
		return "Field is invalid."
	}
}

func Format(ctx *gin.Context, err error) UnifiedErrorResponse {
	// add error to context
	// ctx.Error(err)

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make([]FieldError, len(ve))
		for i, fe := range ve {
			snakeFieldName := strcase.ToSnake(fe.Field())
			out[i] = FieldError{Field: snakeFieldName, Msg: msgForTag(fe.Tag())}
		}
		return UnifiedErrorResponse{
			Success: false,
			Message: "Validation failed",
			Errors:  out,
		}
	}

	// PostgreSQL-specific error handling
	var psqlErr *pgconn.PgError
	if errors.As(err, &psqlErr) {
		return formatPostgresError(psqlErr)
	}

	// Check for pgx.ErrNoRows
	if errors.Is(err, pgx.ErrNoRows) {
		return UnifiedErrorResponse{
			Success: false,
			Message: "No matching record found",
			Errors:  []FieldError{{Field: "root", Msg: "No matching record found"}},
		}
	}

	// Default error handling
	return UnifiedErrorResponse{
		Success: false,
		Message: "An unexpected error occurred",
		Errors:  []FieldError{{Field: "root", Msg: err.Error()}},
	}
}

func formatPostgresError(err *pgconn.PgError) UnifiedErrorResponse {
	var errorMsg string
	var field = "root"
	var message = "Database error" // Default message for PostgreSQL errors

	switch err.Code {
	case "23505":
		errorMsg = "This value is already taken"
		field = extractFieldFromMessage(err.Message)
		message = "Duplicate entry"
	case "1452":
		errorMsg = "Foreign key constraint fails"
		message = "Referential integrity error"
	case "23503":
		errorMsg = "Field cannot be null"
		message = "Missing required field"
	case "3819":
		errorMsg = "Check constraint violation"
		message = "Data integrity violation"
	case "22001":
		errorMsg = "Value too long for field"
		message = "Data too long"
	case "22003":
		errorMsg = "Numeric value out of range"
		message = "Value out of range"
	case "22007":
		errorMsg = "Invalid datetime format"
		message = "Invalid date/time format"
	case "22012":
		errorMsg = "Division by zero"
		message = "Invalid arithmetic operation"
	case "22008":
		errorMsg = "Datetime field overflow"
		message = "Date/time value out of range"
	case "2200B":
		errorMsg = "Escape character error"
		message = "Invalid escape sequence"
	case "2200C":
		errorMsg = "Invalid use of escape character"
		message = "Invalid escape character usage"
	default:
		log.Error().Err(err).Msg("An unexpected database error occurred.")
		return UnifiedErrorResponse{
			Success: false,
			Message: "Unexpected service error. Please try again later.",
			Errors:  []FieldError{{Field: field, Msg: "Unexpected service error. Please try again later."}},
		}
	}

	return UnifiedErrorResponse{
		Success: false,
		Message: message,
		Errors: []FieldError{
			{
				Field: field,
				Msg:   errorMsg,
			},
		},
	}
}

// Extract field name from duplicate entry error like: "Duplicate entry 'foo' for key 'users.email'"
func extractFieldFromMessage(msg string) string {
	// Try to extract the field name from the MySQL error message
	if i := strings.LastIndex(msg, "'"); i > 0 {
		msg = msg[:i]
		if j := strings.LastIndex(msg, "'"); j > 0 {
			parts := strings.Split(msg[j+1:], ".")
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return "field"
}

var (
	ErrRecordNotFound = sql.ErrNoRows
)
