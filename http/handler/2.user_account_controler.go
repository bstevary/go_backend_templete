package handler

import (
	"fmt"
	"net/http"

	"github.com/bstevary/hexagonal/database/model"
	"github.com/bstevary/hexagonal/utils/res"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/gin-gonic/gin"
)

type CodeRequest struct {
	Code string `uri:"code" binding:"required,min=6"`
}

func (u Handler) ActivateUserAccount(c *gin.Context) {
	var req CodeRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	userId, err := u.db.UpdateVerification(c, req.Code)
	if err != nil {
		if err == pgx.ErrNoRows {
			err = fmt.Errorf("invalid or expired verification code")
			c.JSON(http.StatusNotFound, res.Format(c, err))
			return
		}
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	if userId == "" {
		err = fmt.Errorf("invalid or expired verification code")
		c.JSON(http.StatusNotFound, res.Format(c, err))
		return
	}

	err = u.db.UpdateUserStatus(c, model.UpdateUserStatusParams{
		IsActive:      pgtype.Bool{Bool: true, Valid: true},
		ID:            userId,
		AcceptedTerms: pgtype.Bool{Bool: true, Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	c.JSON(http.StatusOK, "")
}
