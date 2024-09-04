package handler

import (
	"phcmis/config"
	"phcmis/databases/persist/db"
	"phcmis/databases/redis/daemon"
	"phcmis/services/auth"

	"github.com/go-redis/cache/v9"
)

type Handler struct {
	db              db.Store
	taskDistributer daemon.TaskDistributor
	tokenizer       auth.TokenGenerator
	config          config.Config
	cache           *cache.Cache
}

func NewHandler(db db.Store, taskDistributer daemon.TaskDistributor, cache *cache.Cache, tokenGenerator auth.TokenGenerator, config config.Config) *Handler {
	return &Handler{
		db:              db,
		taskDistributer: taskDistributer,
		tokenizer:       tokenGenerator,
		config:          config,
		cache:           cache,
	}
}
