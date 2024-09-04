package gin_pgx_err

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrRecordNotFound = pgx.ErrNoRows

const (
	UniqueViolation     = "23505"
	ForeignKeyViolation = "23503"
	NullConstraint      = "23502"
	CheckViolation      = "23514"
	NoResultSet         = "02000"
)

func handleDatabaseError(err *pgconn.PgError) gin.H {
	var errorMsg string
	switch err.Code {
	case UniqueViolation:
		errorMsg = "already taken"
	case ForeignKeyViolation:
		errorMsg = "the provided value is not valid"
	case NullConstraint:
		errorMsg = "field cannot be null"
	case CheckViolation:
		errorMsg = "check constraint violation"
	case NoResultSet:
		errorMsg = "value given not familiar with us"
	default:
		log.Fatal("An unexpected database error occurred ", err)
		return gin.H{"error": []FieldError{{Field: "root", Msg: "An unexpected service error occurred.Issue has been escalated reported. Please try again later."}}}
	}

	return formatError(err.ConstraintName, errorMsg)
}

func formatError(constraintName, msg string) gin.H {
	words := strings.Split(constraintName, "_")
	formattedName := constraintName // Default to the original name in case of unexpected format

	if len(words) >= 3 {
		// Extract the last two words after the constraint name
		formattedName = words[len(words)-3] + "_" + words[len(words)-2]
	}

	return gin.H{
		"error": []FieldError{
			{
				Field: formattedName,
				Msg:   msg,
			},
		},
	}
}
