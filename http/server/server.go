package server

import (
	"context"
	"fmt"
	"net/http"

	"phcmis/config"
	"phcmis/databases/persist/db"
	"phcmis/databases/redis/daemon"
	"phcmis/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	adress string
	router *gin.Engine
}
type ServerDependencies struct {
	Config          config.Config
	DB              db.Store
	Cache           *cache.Cache
	Limiter         *redis_rate.Limiter
	TaskDistributor daemon.TaskDistributor
}

func NewAServer(config config.Config, db db.Store, cache *cache.Cache, limitter *redis_rate.Limiter, taskDistributer daemon.TaskDistributor) (*Server, error) {
	token, err := auth.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot CreateGenerator %w", err)
	}

	router := newRouter(RouterConfig{config, db, token, limitter, taskDistributer, cache})

	return &Server{
		router: router,
		adress: config.HTTPServerAddress,
	}, nil
}

func (server Server) Run(ctx context.Context, waitGroup *errgroup.Group) {
	srv := &http.Server{
		Addr:    server.adress,
		Handler: server.router,
	}
	waitGroup.Go(func() error {
		log.Info().Msgf("server is running at %s", server.adress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("http server failed to start")
			return err
		}

		return nil
	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down server")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("cannot shutdown server")
			return err
		}
		log.Info().Msg("server shutdown successfully")
		return nil
	},
	)
}
