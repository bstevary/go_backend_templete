package handler

import (
	"fmt"
	"net/http"

	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/utils/res"

	"github.com/gin-gonic/gin"
)

type VarifyEmailRequest struct {
	SecretCode string `json:"secret_code" binding:"required"`
}

func (u Handler) ActivateUserAccount(c *gin.Context) {
	var req VarifyEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	txResult, err := u.db.ActivateUserAccountTx(c, db.ActivateUserAccountTxParams{
		SecretCode: req.SecretCode,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	fmt.Println(txResult.User.IsEmailVerified)
	c.JSON(http.StatusOK, txResult.User.IsEmailVerified)
}
