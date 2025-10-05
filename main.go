package main

import (
	"context"
	"embed"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/http/server"
	"github.com/bstevary/hexagonal/jobs"
	"github.com/bstevary/hexagonal/services/mailer"
	"github.com/bstevary/hexagonal/utils/logger"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"

	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"github.com/hibiken/asynq"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

//go:embed database/migrations/*up.sql
var migrations embed.FS

//go:embed emails/*.html
var emails embed.FS

func main() {
	gin.SetMode(gin.ReleaseMode)
	env, err := config.LoadEnv("./")
	if err != nil {
		log.Fatal().Msgf("cannot load configuration: %v", err)
	}
	// initialize the logger
	logger.SetupLogger(logger.LogConfig{LogFilePath: env.LogFilePath, Environment: env.Environment})

	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()

	// initialize the postgress connection pool
	connPool, err := pgxpool.New(ctx, env.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("postgress connection pool creation failed")
	}
	defer connPool.Close()

	// test the postgress database connection
	if err = connPool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgresql")
	}

	// initialize the migration source
	migrationsFiles, err := iofs.New(migrations, "database/migrations")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create migrations source")
	}

	// postgress database migration
	log.Trace().Msg("Applying postgress database migrations")
	m, err := migrate.NewWithSourceInstance("iofs", migrationsFiles, env.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg(" postgress migration failed")
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg(" postgress database migration failed")
	}

	conn := db.NewDatabase(connPool)

	// initialize the redis connection pool
	redisClient := redis.NewClient(&redis.Options{
		Addr:     env.RedisConfig.Host + ":" + env.RedisConfig.Port,
		PoolSize: 5,
		DB:       env.RedisConfig.DB,
	})

	// test the redis connection
	if err = redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Msgf("cannot connect to redis: %v", err)
	}
	// test the redis connection
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatal().Msgf("cannot connect to redis: %v", err)
	}
	log.Debug().Msgf("%v: Redis Connection Successful", pong)

	// initialize the rabbitmq connection
	amqpConn, err := amqp.Dial(env.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to AMQP")
	}
	defer amqpConn.Close()

	rabbitMQ, err := setupRabbitMQChannel(amqpConn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to set up RabbitMQ channel and queue")
	}
	defer rabbitMQ.Close()

	redisOpt := asynq.RedisClientOpt{
		Addr:     env.RedisConfig.Host + ":" + env.RedisConfig.Port,
		PoolSize: 5,
		DB:       env.RedisConfig.DB,
	}

	taskDistributor := jobs.NewRedisTaskDistributor(redisOpt)

	server, err := server.NewHTTPServer(server.ServerDependencies{
		Config:          env,
		DB:              &conn,
		RedisClient:     redisClient,
		TaskDistributor: &taskDistributor,
		S3Uploader:      nil,
		RabbitMQ:        rabbitMQ,
	})
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

	runTaskProcessor(ctx, waitGroup, redisOpt, conn, env)
	if err := waitGroup.Wait(); err != nil {
		log.Fatal().Msgf("error running server: %v", err)
	}
	runTaskScheduler(ctx, waitGroup, redisOpt)
	if err := waitGroup.Wait(); err != nil {
		log.Fatal().Msgf("error running server: %v", err)
	}
}

// setupRabbitMQChannel handles channel creation, exchange declaration, and queue binding.
func setupRabbitMQChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	rabbitMQ, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// 1. Declare Exchange
	err = rabbitMQ.ExchangeDeclare(
		config.TaskExchange, // name
		"direct",            // type
		true,                // durable
		false,               // auto-deleted
		false,               // internal
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Declared RabbitMQ Exchange: %s", config.TaskExchange)

	// 2. Declare Queue
	_, err = rabbitMQ.QueueDeclare(
		config.MessageQueue, // name
		true,                // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Declared RabbitMQ Queue: %s", config.MessageQueue)

	// 3. Bind Queue to Exchange
	err = rabbitMQ.QueueBind(
		config.MessageQueue,
		config.RoutingKey,
		config.TaskExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}
	log.Debug().Msgf("Bound Queue %s to Exchange %s with Key %s", config.MessageQueue, config.TaskExchange, config.RoutingKey)

	// 4. Set QoS (crucial for consumers)
	// This limits the consumer to only 1 unacknowledged message at a time.
	err = rabbitMQ.Qos(
		1,     // prefetchCount
		0,     // prefetchSize
		false, // global
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to set RabbitMQ QoS")
		// Not fatal, but log a warning.
	}

	return rabbitMQ, nil
}

func runTaskProcessor(ctx context.Context, waitGroup *errgroup.Group,
	redisOpt asynq.RedisClientOpt, db db.Database, env *config.Env) {
	mailer := mailer.NewGmailSender(
		env.EmailSenderName,
		env.EmailSenderAddress,
		env.EmailSenderPassword,
	)

	taskProcessor := jobs.NewRedisTaskProcessor(redisOpt, db, mailer, emails)

	waitGroup.Go(func() error {
		if err := taskProcessor.Start(); err != nil {
			log.Error().Err(err).Msg("configQ processor failed to start")
			return err
		}
		return nil

	})

	waitGroup.Go(func() error {
		<-ctx.Done()
		taskProcessor.Shutdown()
		return nil
	})
}

// runTaskScheduler initializes and runs the Asynq scheduler.
func runTaskScheduler(ctx context.Context, waitGroup *errgroup.Group, redisOpt asynq.RedisClientOpt) {
	scheduler := asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{
		// use machine  location
		Location: time.Now().Location(),
	})

	// Run the scheduler within the errgroup for graceful shutdown.
	waitGroup.Go(func() error {
		if err := scheduler.Run(); err != nil {
			log.Error().Err(err).Msg("configQ scheduler failed to run")
			return err
		}
		return nil
	})

	// Listen for the context cancellation to gracefully shut down the scheduler.
	waitGroup.Go(func() error {
		<-ctx.Done()
		scheduler.Shutdown()
		return nil
	})
}
