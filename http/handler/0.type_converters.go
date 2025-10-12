package handler

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func stringToPgText(value string) pgtype.Text {
	if value == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: value, Valid: true}
}

func dateToPgDate(value time.Time) pgtype.Date {
	if value.IsZero() {
		return pgtype.Date{Valid: false}
	}
	return pgtype.Date{Time: value, Valid: true}
}
