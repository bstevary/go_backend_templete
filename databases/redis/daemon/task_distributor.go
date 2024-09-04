package daemon

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskDistributor interface {
	DistributeSendAcitvateAccountInvitationEmail(ctx context.Context, payload *PayloadSendAcitvateAccountInvitationEmail, opts ...asynq.Option) error
}

type RadisTaskDistributor struct {
	client *asynq.Client
}

func NewRadisTaskDistributor(redisOpt asynq.RedisClientOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &RadisTaskDistributor{
		client: client,
	}
}
