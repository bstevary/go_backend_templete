package jobs

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeTaskSendAuthEmail(ctx context.Context, payload *PayloadSendAuthEmail, opts ...asynq.Option) error
}

type RadisTaskDistributor struct {
	client *asynq.Client
}

func NewRedisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RadisTaskDistributor{
		client: client,
	}
}
