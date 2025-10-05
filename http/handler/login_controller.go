package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/database/model"
	"github.com/bstevary/hexagonal/jobs"
	"github.com/guregu/null/v6"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"

	"github.com/bstevary/hexagonal/utils/auth"
	"github.com/bstevary/hexagonal/utils/res"

	"github.com/gin-gonic/gin"
)

type MFAChallengeResponse struct {
	UserID         int64  `json:"user_id"`
	MFAEnabled     bool   `json:"mfa_enabled"`
	Message        string `json:"message"`
	MFAChallengeID string `json:"mfa_challenge_id,omitempty"`
}

type LoginUserResp struct {
	SessionID            string    `json:"session_id"`
	AccessToken          string    `json:"access_token"`
	RefreshToken         string    `json:"refresh_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
	ActiveBranch         int64     `json:"active_branch"`
	ActiveOrganisation   int64     `json:"active_organisation"`
	Branches             []int64   `json:"branches"`
	Organisations        []int64   `json:"organisations"`
	Permissions          []string  `json:"permissions"`
	User                 User      `json:"user"`
}
type User struct {
	UserID    int64  `json:"user_id" `
	FirstName string `json:"first_name" `
	LastName  string `json:"last_name" `
	Username  string `json:"username" `
	Title     string `json:"title" `
	Gender    string `json:"gender" `
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=4,max=50"`
}

func (u Handler) UserLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	user, err := u.db.SelectUserByEmail(c, req.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to select user by email")
		err = fmt.Errorf("Email or Password is not valid")
		c.JSON(http.StatusUnauthorized, res.Format(c, err))
		return
	}

	err = auth.CheckPassword(user.Password, req.Password)
	if err != nil {
		log.Error().Err(err).Msg("failed varify password")
		err = fmt.Errorf("%s  or Password is not valid", UserInfo)
		c.JSON(http.StatusUnauthorized, res.Format(c, err))
		return
	}
	if !user.IsVerified {
		err = fmt.Errorf(" your account is not verified yet, please check your email to verify your account")
		c.JSON(http.StatusUnauthorized, res.Format(c, err))
		return
	}

	if user.IsLocked {
		err = fmt.Errorf(" your account is locked, please contact the admin")
		c.JSON(http.StatusUnauthorized, res.Format(c, err))
		return
	}

	if !user.IsActive {
		err = fmt.Errorf(" your account is disabled, please contact the admin")
		c.JSON(http.StatusUnauthorized, res.Format(c, err))
		return
	}

	// --- MFA Handling ---
	if user.MfaEnabled {
		// Dispatch an jobs task to send the email/SMS
		taskPayload := &jobs.PayloadSendAuthEmail{
			UserID: user.ID,
			Type:   "default",
		}
		opts := []asynq.Option{
			asynq.ProcessIn(2 * time.Second),
			asynq.Queue(jobs.CriticalQueue),
			asynq.MaxRetry(4),
		}

		if err := u.taskDistributer.DistributeTaskSendAuthEmail(c, taskPayload, opts...); err != nil {
			log.Error().Err(err).Msg("failed to enqueue MFA email task")
			c.JSON(http.StatusInternalServerError, res.Format(c, fmt.Errorf("failed to initiate MFA, please try again")))
			return
		}

		c.JSON(http.StatusOK, MFAChallengeResponse{
			UserID:     user.ID,
			MFAEnabled: true,
			Message:    "MFA challenge initiated. Please provide the code sent to your registered MFA device/email.",
		})
		return
	}
	// --- End MFA Handling ---

	userRoles, err := u.db.GetUserRolesByBranch(c, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	raw, err := u.db.GetUserOrganisationIds(c, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	Organisations, activeOrganisation := getDefaultOrganisation(raw)

	branchIDs, activeBranch, Permissions := getPermissionsForSelectedBranch(userRoles, 0, false)

	accessToken, accessPayload, err := u.tokenizer.CreateToken(
		user.ID, Permissions, u.config.AccessTokenDuration, c.ClientIP(), activeBranch, branchIDs, activeOrganisation, Organisations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	refreshToken, refreshPayload, err := u.tokenizer.CreateToken(
		user.ID, Permissions, u.config.RefreshTokenDuration, c.ClientIP(), activeBranch, branchIDs, activeOrganisation, Organisations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
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

	err = u.db.CreateSession(c, model.CreateSessionParams{
		ID:           refreshPayload.ID.String(),
		RefreshToken: refreshToken,
		UserAgent:    c.Request.UserAgent(),
		ClientIp:     c.ClientIP(),
		Expiry:       refreshPayload.ExpiredAt,
		Staff:        user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	rsp := LoginUserResp{
		SessionID:            refreshPayload.ID.String(),
		AccessToken:          accessToken,
		ActiveBranch:         activeBranch,
		Permissions:          Permissions,
		Branches:             branchIDs,
		RefreshToken:         refreshToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
		Organisations:        Organisations,
		ActiveOrganisation:   activeOrganisation,
		User: User{
			UserID:    user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
			Title:     user.Title,
			Gender:    user.Gender,
		},
	}
	c.JSON(http.StatusOK, rsp)
}

type VerifyMFARequest struct {
	UserID         int64  `json:"user_id" binding:"required"`
	MFAChallengeID string `json:"mfa_challenge_id,omitempty"`
	MFACode        string `json:"mfa_code" binding:"required,len=6"`
}

func (u Handler) MFAChallenge(c *gin.Context) {
	var req VerifyMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	user_Id, err := u.db.GetVarificationUser(c, req.MFACode)
	if err != nil {
		log.Error().Err(err).Int64("UserID", req.UserID).Msg("failed to update MFA verification")
		c.JSON(http.StatusBadRequest, res.Format(c, fmt.Errorf("failed to pass MFA challenge")))
		return

	}
	if user_Id != req.UserID {
		log.Error().Err(err).Int64("UserID", req.UserID).Msg("failed to update MFA verification")
		c.JSON(http.StatusBadRequest, res.Format(c, fmt.Errorf("failed to pass MFA challenge")))
		return

	}
	err = u.db.UpdateVerification(c, req.MFACode)
	if err != nil {
		log.Error().Err(err).Int64("UserID", req.UserID).Msg("failed to update MFA verification")
		c.JSON(http.StatusBadRequest, res.Format(c, fmt.Errorf("failed to pass MFA challenge")))
		return

	}

	user, err := u.db.GetUser(c, req.UserID)
	if err != nil {
		log.Error().Err(err).Int64("UserID", req.UserID).Msg("failed to select user by ID during MFA verification")
		c.JSON(http.StatusInternalServerError, res.Format(c, fmt.Errorf("internal server error during MFA verification")))
		return
	}
	if !user.MfaEnabled {
		c.JSON(http.StatusBadRequest, res.Format(c, fmt.Errorf("MFA is not enabled for this user")))
		return
	}

	userRoles, err := u.db.GetUserRolesByBranch(c, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	raw, err := u.db.GetUserOrganisationIds(c, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	Organisations, activeOrganisation := getDefaultOrganisation(raw)

	branchIDs, activeBranch, Permissions := getPermissionsForSelectedBranch(userRoles, 0, false)

	accessToken, accessPayload, err := u.tokenizer.CreateToken(
		user.ID, Permissions, u.config.AccessTokenDuration, c.ClientIP(), activeBranch, branchIDs, activeOrganisation, Organisations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	refreshToken, refreshPayload, err := u.tokenizer.CreateToken(
		user.ID, Permissions, u.config.RefreshTokenDuration, c.ClientIP(), activeBranch, branchIDs, activeOrganisation, Organisations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  refreshPayload.ExpiredAt,
		Path:     "/",
		Domain:   u.config.Domain,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	err = u.db.CreateSession(c, model.CreateSessionParams{
		ID:           refreshPayload.ID.String(),
		RefreshToken: refreshToken,
		UserAgent:    c.Request.UserAgent(),
		ClientIp:     c.ClientIP(),
		Expiry:       refreshPayload.ExpiredAt,
		Staff:        user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	rsp := LoginUserResp{
		SessionID:            refreshPayload.ID.String(),
		AccessToken:          accessToken,
		ActiveBranch:         activeBranch,
		Permissions:          Permissions,
		Branches:             branchIDs,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
		Organisations:        Organisations,
		User: User{
			UserID:    user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Username:  user.Username,
			Title:     user.Title,
			Gender:    user.Gender,
		},
	}
	c.JSON(http.StatusOK, rsp)
}

type UserEmail struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetUserRequest struct {
	Otp             string `json:"otp" binding:"required"`
	Password        string `json:"password" binding:"required,min=8,max=20"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

func (u Handler) ResetAccountPassword(c *gin.Context) {
	var req ResetUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	hashPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}
	err = u.db.ActivateUserAccountTx(c, db.ActivateUserAccountTxParams{
		SecretCode:     req.Otp,
		HashedPassword: hashPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	c.JSON(http.StatusOK, "")
}

func (u Handler) ForgotPassword(c *gin.Context) {
	var req UserEmail
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}
	user, err := u.db.SelectUserByEmail(c, req.Email)
	if err != nil {
		log.Error().Err(err).Msg("failed to select user")
		c.JSON(http.StatusOK, "")
		return
	}

	taskPayload := &jobs.PayloadSendAuthEmail{
		UserID: user.ID,
		Type:   "Password",
	}

	opts := []asynq.Option{
		asynq.ProcessIn(10 * time.Second),
		asynq.Queue(jobs.CriticalQueue),
		asynq.MaxRetry(4),
	}

	if err := u.taskDistributer.DistributeTaskSendAuthEmail(c, taskPayload, opts...); err != nil {
		log.Error().Err(err).Msg("failed to distribute forgot password  email task")
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	c.JSON(http.StatusOK, "")
}

type ResetPasswordRequest struct {
	Password        string `json:"password" binding:"required,min=8,max=20"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}
type OTPQuery struct {
	Otp string `uri:"otp" binding:"required"`
}

func (u Handler) ResetPassword(c *gin.Context) {
	var qry OTPQuery
	if err := c.ShouldBindUri(&qry); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	hashPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}
	err = u.db.ActivateUserAccountTx(c, db.ActivateUserAccountTxParams{
		SecretCode:     qry.Otp,
		HashedPassword: hashPassword,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	c.JSON(http.StatusOK, "")
}

type ChangePasswordRequest struct {
	Password        string `json:"password" binding:"required,min=8,max=20"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=20"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

func (u Handler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	payload := c.MustGet(auth.AuthKey).(*auth.Payload)
	user, err := u.db.SelectUser(c, payload.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	err = auth.CheckPassword(user.Password, req.Password)
	if err != nil {
		err = fmt.Errorf("wrong password")
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	hashPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}
	err = u.db.UpdateUser(c, model.UpdateUserParams{
		ID:       payload.UserID,
		Password: null.StringFrom(hashPassword),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	c.JSON(http.StatusOK, "")
}

func (u Handler) Logout(c *gin.Context) {
	cookie, err := c.Request.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	payload, err := u.tokenizer.ValidateToken(cookie.Value)
	if err != nil {
		c.JSON(http.StatusUnauthorized, res.Format(c, err))
		return
	}

	err = u.db.DeleteSession(c, payload.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
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

	c.JSON(http.StatusOK, "")
}
