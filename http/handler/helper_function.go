package handler

import (
	"time"

	"phcmis/databases/persist/model"

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

type ListUsersResponse struct {
	Users      []model.ListUsersRow `json:"users"`
	NextCursor string               `json:"next_cursor"`
	RowCount   int64                `json:"row_count"`
}
