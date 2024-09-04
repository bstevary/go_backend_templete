package daemon

import (
	"context"

	"phcmis/databases/persist/db"
	"phcmis/services/gmail"

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
	ProcessSendAcitvateAccountInvitationEmailTask(ctx context.Context, task *asynq.Task) error
}

type RadisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
	mailer gmail.EmailSender
}

func NewRadisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store, mailer gmail.EmailSender) TaskProcessor {
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
	return &RadisTaskProcessor{
		server: server,
		store:  store,
		mailer: mailer,
	}
}

func (r *RadisTaskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskSendAcitvateAccountEmail, r.ProcessSendAcitvateAccountInvitationEmailTask)
	return r.server.Start(mux)
}

func (r *RadisTaskProcessor) Shutdown() {
	r.server.Shutdown()
}
