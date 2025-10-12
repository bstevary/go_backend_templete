package handler

import (
	"github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/jobs"
	"github.com/bstevary/hexagonal/services/S3"
	"github.com/bstevary/hexagonal/utils/auth"
	"github.com/bstevary/hexagonal/utils/id"
	"github.com/go-redis/cache/v9"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler struct {
	db              db.Database
	taskDistributer jobs.TaskDistributor
	tokenizer       auth.TokenGenerator
	config          *config.Env
	cache           *cache.Cache
	rabbitMQ        *amqp.Channel
	S3              S3.S3Uploader
	IdGen           *id.IDGenerator
}

func NewHandler(db db.Database, taskDistributer jobs.TaskDistributor,
	cache *cache.Cache, tokenGenerator auth.TokenGenerator, config *config.Env,
	rabbitMQ *amqp.Channel, s3 S3.S3Uploader, IdGen *id.IDGenerator) *Handler {
	return &Handler{
		db:              db,
		taskDistributer: taskDistributer,
		tokenizer:       tokenGenerator,
		config:          config,
		cache:           cache,
		rabbitMQ:        rabbitMQ,
		S3:              s3,
		IdGen:           IdGen,
	}
}

type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type PaginatedRequest struct {
	Page int32 `form:"page" json:"page" binding:"required,min=1"`
	Size int32 `form:"size" json:"size" binding:"required,min=1,max=200"`
}

type Meta struct {
	Page  int32 `json:"page"`
	Size  int32 `json:"size"`
	Total int64 `json:"total"`
}

type PaginatedResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    Meta   `json:"meta"`
}
