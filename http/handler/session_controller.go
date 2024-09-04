package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"phcmis/services/gin_pgx_err"

	"github.com/gin-gonic/gin"
)

func (u Handler) RenewAcessToken(c *gin.Context) {
	cookie, err := c.Request.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	refreshPayload, err := u.tokenizer.ValidateToken(cookie.Value)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	if refreshPayload.Valid() != nil {
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	if refreshPayload.ClientIP != c.ClientIP() {
		err := errors.New("token bound to a different IP address")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	sesseion, err := u.db.GetSession(c, refreshPayload.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}
	if sesseion.IsBlocked {
		err = fmt.Errorf("session is blocked")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	if sesseion.Email != refreshPayload.Email {
		err = fmt.Errorf("incorect sesseion user")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	if sesseion.RefreshToken != cookie.Value {
		err = fmt.Errorf("incorect sesseion refresh token")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}
	if sesseion.UserAgent != c.Request.UserAgent() {
		err = fmt.Errorf("incorect sesseion user agent")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}
	if sesseion.ClientIp != c.ClientIP() {
		err = fmt.Errorf("incorect sesseion client ip")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}
	if time.Now().After(sesseion.ExpiresAt) {
		err = fmt.Errorf("expired sesseion")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	accessToken, accessPayload, err := u.tokenizer.CreateToken(
		refreshPayload.Email,
		u.config.AccessTokenDuration, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	user, err := u.db.SelectUserByEmail(c, refreshPayload.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	rsp := LoginUserResp{
		SessionID:            sesseion.ID,
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
		User: AuthUserResponse{
			UserID:    user.UserID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Gender:    user.Gender,
			Email:     user.Email},
	}

	c.JSON(http.StatusOK, rsp)
}
