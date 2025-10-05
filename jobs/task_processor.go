package jobs

import (
	"context"
	"embed"
	"html/template"

	"github.com/bstevary/hexagonal/database/db"
	"github.com/bstevary/hexagonal/services/mailer"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	CriticalQueue = "critical"
	DefaultQueue  = "default"
)

type TaskProcessor interface {
	Start() error
	Shutdown()
	ProcessSendAuthEmailTask(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server   *asynq.Server
	db       db.Database
	mailer   mailer.MailSender
	template *template.Template
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, db db.Database,
	mailer mailer.MailSender, emails embed.FS) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)
	server := asynq.NewServer(redisOpt, asynq.Config{
		Queues: map[string]int{
			CriticalQueue: 10,
			DefaultQueue:  5,
		},
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			log.Error().Err(err).Str("type", task.Type()).
				Bytes("payload", task.Payload()).Msg("process task failed")
		}),
		Logger: logger,
	})

	template, err := template.New("emails").ParseFS(emails, "emails/*.html")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create emails source")
	}
	return &RedisTaskProcessor{
		server:   server,
		db:       db,
		mailer:   mailer,
		template: template,
	}
}

func (r *RedisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(SendAuthEmailTask, r.ProcessSendAuthEmailTask)
	return r.server.Start(mux)
}

func (r *RedisTaskProcessor) Shutdown() {
	r.server.Shutdown()
}
