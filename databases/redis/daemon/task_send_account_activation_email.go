package daemon

import (
	"context"
	"encoding/json"
	"fmt"

	"phcmis/databases/persist/model"
	"phcmis/services/gin_pgx_err"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const TaskSendAcitvateAccountEmail = "task:send_acitvate_account_invitation_email"

type PayloadSendAcitvateAccountInvitationEmail struct {
	Email      string `json:"email"`
	SecretCode string `json:"secret_code"`
}

func (d *RadisTaskDistributor) DistributeSendAcitvateAccountInvitationEmail(ctx context.Context, payload *PayloadSendAcitvateAccountInvitationEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(TaskSendAcitvateAccountEmail, jsonPayload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("Enqueued task")

	return nil
}

func (processor *RadisTaskProcessor) ProcessSendAcitvateAccountInvitationEmailTask(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendAcitvateAccountInvitationEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", asynq.SkipRetry)
	}

	user, err := processor.store.SelectUserByEmail(ctx, payload.Email)
	if err != nil {
		if err == gin_pgx_err.ErrRecordNotFound {
			return fmt.Errorf("user %q not found: %w", payload.Email, asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user %q: %w", payload.Email, err)
	}

	log.Info().Str("type", task.Type()).Str("user_name", user.FirstName).Msg("Processing task")

	if user.Email == "" {
		return fmt.Errorf("user %q has no email: %w", payload.Email, asynq.SkipRetry)
	}

	verifyEmail, err := processor.store.CreateActivateAccountEmail(ctx, model.CreateActivateAccountEmailParams{
		UserID:     user.UserID,
		Email:      user.Email,
		SecretCode: payload.SecretCode,
	})
	if err != nil {
		return fmt.Errorf("failed to create account activation email: %w", err)
	}

	subject := "Welcome to phcmis!"
	// TODO: replace this URL with an environment variable that points to a front-end page
	verifyUrl := fmt.Sprintf("http://localhost:8080/activate?&secret_code=%s",
		verifyEmail.SecretCode)
	content := fmt.Sprintf(`Hello  %s  %s,<br/>
	We have Created you in the system.  To access your Account  kindly activate with these !<br/>
	SecretCode: %s<br/>
	Primary Health Care Id (Phc_ID): %s<br/>
	Welcome onbord.<br/>
	Welcome to the system!<br/>
	 
	Please <a href="%s">click here</a> to Acknowledge and Accept the Role.<br/>
	`, user.FirstName, user.LastName, verifyEmail.SecretCode, payload.Email, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to send verify email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	// ... send email
	return nil
}
