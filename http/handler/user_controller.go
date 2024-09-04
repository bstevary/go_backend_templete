package handler

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"time"

	"phcmis/databases/persist/db"
	"phcmis/databases/persist/model"
	"phcmis/databases/redis/daemon"
	"phcmis/services/auth"
	"phcmis/services/gin_pgx_err"
	"phcmis/services/phc"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type CreateUserRequest struct {
	FirstName   string    `json:"first_name" binding:"required,alphanum,min=3,max=20"`
	LastName    string    `json:"last_name" binding:"required,alphanum,min=3,max=20"`
	Gender      string    `json:"gender" binding:"required,oneof= male binary female"`
	Email       string    `json:"email" binding:"required,email,min=10,max=60"`
	DateOfBirth time.Time `json:"date_of_birth" binding:"required" time_format:"2006-01-02"`
	Password    string    `json:"password" binding:"required,min=8,max=20"`
}

func (u Handler) CreateUserAccount(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	HashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: model.CreateUserParams{
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			Email:          req.Email,
			HashedPassword: HashedPassword,
			Gender:         req.Gender,
			DateOfBirth:    req.DateOfBirth,
		},
		AfterCreateUser: func(user model.CreateUserRow) error {
			//  send email verification
			taskPayload := &daemon.PayloadSendAcitvateAccountInvitationEmail{
				Email:      user.Email,
				SecretCode: phc.GeneratePHCID(),
			}

			opts := []asynq.Option{
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(daemon.CriticalQueue),
			}

			return u.taskDistributer.DistributeSendAcitvateAccountInvitationEmail(c, taskPayload, opts...)
		},
	}

	txResult, err := u.db.CreateUserTx(c, arg)
	if err != nil {

		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, txResult.User)
}

type UserRequestQuery struct {
	Email string `uri:"email" binding:"required,alphanum,min=10,max=15"`
}

type GetUserResponse struct {
	UserID      int64     `json:"user_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender"`
}

func (u Handler) GetUser(c *gin.Context) {
	var req UserRequestQuery
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}
	user, err := u.db.SelectUserByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}
	resp := GetUserResponse{
		UserID:      user.UserID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		DateOfBirth: user.DateOfBirth,
		Gender:      user.Gender,
	}

	c.JSON(http.StatusOK, resp)
}

type UpdateUserRequest struct {
	FirstName   string    `json:"first_name" binding:"omitempty,alphanum,min=3,max=20"`
	DateOfBirth time.Time `json:"date_of_birth" binding:"required" time_format:"2006-01-02"`
	Surname     string    `json:"surname,omitempty" binding:"omitempty,min=3,max=20" `
	LastName    string    `json:"last_name" binding:"omitempty,alphanum,min=3,max=20"`
	Gender      string    `json:"gender" binding:"omitempty,oneof= male binary female"`
}

func (u Handler) UpdateUser(c *gin.Context) {
	var reqQ UserRequestQuery
	if err := c.ShouldBindUri(&reqQ); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	var req UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	arg := model.UpdateUserParams{
		FirstName:   stringToPgText(req.FirstName),
		LastName:    stringToPgText(req.LastName),
		DateOfBirth: dateToPgDate(req.DateOfBirth),
		Gender:      stringToPgText(req.Gender),
		Email:       reqQ.Email,
	}
	user, err := u.db.UpdateUser(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, user)
}

func (u Handler) DeleteUser(c *gin.Context) {
	var req UserRequestQuery
	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}
	PHC_ID, err := u.db.DeleteUser(c, req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully", "PHC_ID": PHC_ID})
}

type ListUsersRequest struct {
	NextCursor string `form:"next_cursor"`
	PageSize   int32  `form:"page_size" binding:"required"`
}

type ListAllUsersResponse struct {
	Users      []model.ListUsersRow `json:"users"`
	NextCursor string               `json:"next_cursor"`
	RowCount   int64                `json:"row_count"`
}

func (u Handler) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}
	var arg model.ListUsersParams

	if req.NextCursor != "" {
		decodedCursor, err := base64.RawURLEncoding.DecodeString(req.NextCursor)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
			return
		}

		decodedInt, err := strconv.ParseInt(string(decodedCursor), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
			return
		}

		arg = model.ListUsersParams{
			UserID: decodedInt,
			Limit:  req.PageSize + 1,
		}
	} else {
		arg = model.ListUsersParams{
			Limit: req.PageSize + 1,
		}
	}

	users, err := u.db.ListUsers(c, arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin_pgx_err.ErrorResponse(err))
		return
	}

	var resp ListAllUsersResponse
	if len(users) > int(req.PageSize) {
		users = users[:len(users)-1]
		resp.NextCursor = base64.StdEncoding.EncodeToString([]byte(strconv.FormatInt(users[len(users)-1].UserID, 10)))
	}

	resp.RowCount, err = u.db.CountUsers(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin_pgx_err.ErrorResponse(err))
		return
	}

	resp.Users = users

	c.JSON(http.StatusOK, resp)
}
