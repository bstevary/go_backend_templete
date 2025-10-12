package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/database/model"
	"github.com/bstevary/hexagonal/jobs"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/bstevary/hexagonal/utils/auth"
	"github.com/bstevary/hexagonal/utils/res"

	"github.com/gin-gonic/gin"
	"github.com/hibiken/asynq"
)

type CreateUserRequest struct {
	FirstName       string `json:"first_name" binding:"required,alphanum,min=3,max=20"`
	LastName        string `json:"last_name" binding:"required,alphanum,min=3,max=20"`
	MiddleName      string `json:"middle_name" binding:"omitempty,alphanum,max=20"`
	OtherName       string `json:"other_name" binding:"omitempty,alphanum,max=20"`
	Email           string `json:"email" binding:"required,email,min=10,max=60"`
	Terms           bool   `json:"terms" binding:"required"`
	Password        string `json:"password" binding:"required,min=8,max=20"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`

	// DateOfBirth time.Time `json:"date_of_birth" binding:"required" time_format:"2006-01-02"`
}

func (u Handler) CreateUserAccount(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error().Err(err).Str("Method", "CreateUserAccount").Msg("failed to bind json CreateUserRequest")
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	HashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Error().Err(err).Str("Method", "CreateUserAccount").Msg("failed to hash password")
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	Id, err := u.IdGen.GetUserID()
	if err != nil {
		log.Error().Err(err).Str("Method", "CreateUserAccount").Msg("failed to generate user id")
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	arg := db.CreateUserTxParams{
		CreateUserParams: model.CreateUserParams{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Password:  HashedPassword,
			ID:        Id,
		},
		AfterCreateUser: func(UserID string) error {
			//  send email verification
			taskPayload := &jobs.PayloadSendAuthEmail{
				UserID: UserID,
				Type:   jobs.NewAccountActivation,
			}

			opts := []asynq.Option{
				asynq.ProcessIn(10 * time.Second),
				asynq.Queue(jobs.CriticalQueue),
			}

			return u.taskDistributer.DistributeTaskSendAuthEmail(c, taskPayload, opts...)
		},
	}

	err = u.db.CreateUserTx(c, arg)
	if err != nil {
		log.Error().Err(err).Str("Method", "CreateUserAccount").Msg("failed to create user transaction")
		if err == pgx.ErrNoRows {
			err = fmt.Errorf("email already exists")
			c.JSON(http.StatusBadRequest, res.Format(c, err))
			return
		}
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	c.JSON(http.StatusOK, "")
}

func (u Handler) ListUsers(c *gin.Context) {
	var req PaginatedRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, res.Format(c, err))
		return
	}

	users, err := u.db.ListUsers(c, model.ListUsersParams{
		Limit:  req.Size,
		Offset: (req.Page - 1) * req.Size,
	})
	if err != nil {
		log.Error().Err(err).Str("Method", "ListUsers").Msg("failed to list users")
		c.JSON(http.StatusInternalServerError, res.Format(c, err))
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: "List of users",
		Data:    users,
	})
}
