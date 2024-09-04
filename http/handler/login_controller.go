package handler

import (
	"fmt"
	"net/http"
	"time"

	"phcmis/databases/persist/model"
	"phcmis/services/auth"
	"phcmis/services/gin_pgx_err"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,min=3,max=80"`
	Password string `json:"password" binding:"required,min=4,max=50"`
}

type LoginUserResp struct {
	SessionID            uuid.UUID        `json:"session_id"`
	AccessToken          string           `json:"access_token"`
	AccessTokenExpiresAt time.Time        `json:"access_token_expires_at"`
	User                 AuthUserResponse `json:"user_info"`
}
type AuthUserResponse struct {
	UserID    int64  `json:"user_id" `
	FirstName string `json:"first_name" `
	LastName  string `json:"last_name" `
	Email     string `json:"username" `
	Gender    string `json:"gender" `
}

func (u Handler) UserLogin(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	user, err := u.db.SelectUserByEmail(c, req.Email)
	if err != nil {
		if err == gin_pgx_err.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(fmt.Errorf("invalid email or password")))
			return
		}
		c.JSON(http.StatusUnprocessableEntity, gin_pgx_err.ErrorResponse(err))
		return
	}

	err = auth.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(fmt.Errorf("invalid email or password")))
		return
	}

	// Use isActiveAccount and proceed only if the account is active
	if !user.IsAccountActive {
		err = fmt.Errorf("your account activation is pending. check your email for the activation link")
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	accessToken, accessPayload, err := u.tokenizer.CreateToken(req.Email, u.config.AccessTokenDuration, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}
	refreshToken, refreshPayload, err := u.tokenizer.CreateToken(req.Email, u.config.AccessTokenDuration, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	// Set the refresh token in an HTTPS-only cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  refreshPayload.ExpiredAt,
		Path:     "/",
		Domain:   u.config.Domain,
		HttpOnly: true,
		Secure:   true,
	})

	sesseion, err := u.db.CreateSession(c, model.CreateSessionParams{
		ID:           refreshPayload.ID,
		Email:        req.Email,
		RefreshToken: refreshToken,
		UserAgent:    c.Request.UserAgent(),
		ClientIp:     c.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
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
			Email:     user.Email,
			Gender:    user.Gender,
		},
	}
	c.JSON(http.StatusOK, rsp)
}

func (u Handler) ForgotPassword(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (u Handler) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (u Handler) ChangePassword(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (u Handler) Logout(c *gin.Context) {
	cookie, err := c.Request.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	payload, err := u.tokenizer.ValidateToken(cookie.Value)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin_pgx_err.ErrorResponse(err))
		return
	}

	err = u.db.DeleteSession(c, payload.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	// Delete the refresh token cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   0,
		Path:     "/",
		Domain:   u.config.Domain,
		HttpOnly: true,
		Secure:   true,
	})

	c.JSON(http.StatusOK, "successfully logged out")
}
