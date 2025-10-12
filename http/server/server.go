package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/http/handler"
	"github.com/bstevary/hexagonal/jobs"
	"github.com/bstevary/hexagonal/services/S3"
	"github.com/bstevary/hexagonal/utils/auth"
	"github.com/bstevary/hexagonal/utils/id"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	adress string
	router *gin.Engine
}
type ServerDependencies struct {
	Config          *config.Env
	DB              db.Database
	RedisClient     *redis.Client
	TaskDistributor jobs.TaskDistributor
	S3Uploader      S3.S3Uploader
	RabbitMQ        *amqp.Channel
	IdGen           *id.IDGenerator
}

func NewHTTPServer(arg ServerDependencies) (*Server, error) {
	token, err := auth.NewPasetoMaker(arg.Config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot CreateGenerator %w", err)
	}
	radisCache := cache.New(&cache.Options{
		Redis:      arg.RedisClient,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	handler := handler.NewHandler(arg.DB, arg.TaskDistributor, radisCache,
		token, arg.Config, arg.RabbitMQ, arg.S3Uploader, arg.IdGen)

	router := newRouter(routerConfig{
		handler: handler,
		env:     arg.Config,
		token:   token,
		limiter: redis_rate.NewLimiter(arg.RedisClient),
	})

	return &Server{
		router: router,
		adress: arg.Config.ServerAddress,
	}, nil
}

func (server Server) Run(ctx context.Context, waitGroup *errgroup.Group) {
	srv := &http.Server{
		Addr:    server.adress,
		Handler: server.router,
	}

	waitGroup.Go(func() error {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("http server failed to start")
			return err
		}
		log.Debug().Msgf("http server  is listerning  at  %s", server.adress)

		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("cannot shutdown server")
			return err
		}
		return nil
	},
	)
}
