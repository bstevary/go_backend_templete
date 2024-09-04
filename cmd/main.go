package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"phcmis/config"
	"phcmis/databases/persist/db"
	"phcmis/databases/redis/daemon"
	"phcmis/http/server"
	"phcmis/services/gmail"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/go-redis/redis_rate/v10"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

//@title PHCMIS API
//@version 1.0
//@description This is the API for the PHCMIS system

//@host localhost:8080
//@BasePath /
//@schemes http
//@produce json
//@consumes json

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	config, err := config.LoadConfig("../")
	// mpsea trasaction
	if err != nil {
		log.Fatal().Msgf("cannot load configuration: %v", err)
	}
	log.Logger = zerolog.New(gin.DefaultWriter).With().Timestamp().Logger()
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		gin.SetMode(gin.DebugMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()
	connPool, err := pgxpool.New(ctx, config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot connect to database: %v", err)
	}

	conn := db.NewStore(connPool)

	// Run Migrations
	runMigrations(&config)

	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}

	taskDistributor := daemon.NewRadisTaskDistributor(redisOpt)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddress,
		PoolSize: 20,
	})

	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Msgf("cannot connect to redis: %v", err)
	}
	limiter := redis_rate.NewLimiter(redisClient)

	log.Info().Msgf("%v: Redis Connection Successful", pong)
	cache := cache.New(&cache.Options{
		Redis:      redisClient,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	server, err := server.NewAServer(config, conn, cache, limiter, taskDistributor)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %v", err)
	}

	waitGroup, ctx := errgroup.WithContext(ctx)
	server.Run(ctx, waitGroup)

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down Redis client gracefully")
		if err := redisClient.Close(); err != nil {
			log.Error().Msgf("error closing Redis client: %v", err)
			return err
		}
		log.Info().Msg("Redis client shutdown successful")
		return nil
	})

	runTaskProcessor(ctx, waitGroup, redisOpt, conn, config)
	if err := waitGroup.Wait(); err != nil {
		log.Fatal().Msgf("error running server: %v", err)
	}
}

func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group, redisOpt asynq.RedisClientOpt, store db.Store, config config.Config) {
	mailer := gmail.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	taskProcessor := daemon.NewRadisTaskProcessor(redisOpt, store, mailer)
	if err := taskProcessor.Start(); err != nil {
		log.Fatal().Err(err).Msg("error from wait group")
	}

	waitGroup.Go(func() error {
		<-ctx.Done()
		log.Info().Msg("shutting down task processor gracefully")
		taskProcessor.Shutdown()
		log.Info().Msg("task processor shutdown successful")
		return nil
	})
}

func runMigrations(config *config.Config) {
	m, err := migrate.New(config.MigrationURL, config.DBSource)
	if err != nil {
		log.Fatal().Msgf("cannot create migration instance: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msgf("cannot run migration: %v", err)
	}
	log.Info().Msg("Migration successful")
}
