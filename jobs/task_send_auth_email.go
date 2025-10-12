package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"time"

	"github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/database/model"
	"github.com/bstevary/hexagonal/services/mailer"
	"github.com/bstevary/hexagonal/utils/id"
	"github.com/bstevary/hexagonal/utils/res"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog/log"
)

const SendAuthEmailTask = "task:send_auth_email"

type AuthMailContent struct {
	Title          string
	Name           string
	Message        template.HTML
	OTP            string
	Action         string
	Link           string
	Concern        string
	Year           int
	CompanyName    string
	CompanyURL     string
	CompanyAddress string
	SupportEmail   string
}

// AuthMail encapsulates the content and recipients for an email.
type AuthMail struct {
	Content AuthMailContent
	To      []string
}

const (
	NewAccountActivation  = "New"
	PasswordReset         = "Password"
	AuthorizeLoginAttempt = "Login"
)

type PayloadSendAuthEmail struct {
	UserID string `json:"user_id"`
	Type   string `json:"type"`
}

func (d *RedisTaskDistributor) DistributeTaskSendAuthEmail(ctx context.Context, payload *PayloadSendAuthEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}
	task := asynq.NewTask(SendAuthEmailTask, jsonPayload, opts...)

	info, err := d.client.EnqueueContext(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to enqueue task: %w", err)
	}
	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).Str("queue", info.Queue).Int("max_retry", info.MaxRetry).Msg("Enqueued task")

	return nil
}

func (prc *RedisTaskProcessor) ProcessSendAuthEmailTask(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendAuthEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %w", asynq.SkipRetry)
	}

	user, err := prc.db.GetUser(ctx, payload.UserID)
	if err != nil {
		if err == res.ErrRecordNotFound {
			return fmt.Errorf("user %q not found: %w", payload.UserID, asynq.SkipRetry)
		}
		return fmt.Errorf("failed to select user: %w", err)
	}

	otpCode, err := id.GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	err = prc.db.CreateVerification(ctx, model.CreateVerificationParams{
		Otp:     otpCode,
		UserID:  user.ID,
		Channel: "EMAIL",
	})
	if err != nil {
		return fmt.Errorf("failed to create verification record: %w", err)
	}

	// Define message details based on payload.Type
	messageDetails := map[string]struct {
		Title   string
		Message string
		Action  string
		Concern string
	}{
		NewAccountActivation: {
			Title:   "Activate Your Account",
			Message: `Welcome! Please use the One-Time Password (OTP) below to activate your account.`,
			Action:  "Verify Now",
			Concern: "If you didn’t sign up, please ignore this email.",
		},

		PasswordReset: {
			Title:   "Reset Password Request",
			Message: `You requested a password reset. Use the One-Time Password (OTP) below to reset your password.`,
			Action:  "Reset Password",
			Concern: `If this wasn't you, it is safe to ignore this email.`,
		},
		"default": { // Fallback for unmatched types
			Title:   "Authorize Your Account Login",
			Message: `A login attempt was detected. Use the One-Time Password (OTP) below to verify your identity.`,
			Action:  "Authorize Login",
			Concern: `If this wasn't you, your account remains secure.`,
		},
	}

	// Retrieve specific details, falling back to "default" if type is not found
	details, ok := messageDetails[payload.Type]
	if !ok {
		details = messageDetails["default"]
	}

	// Construct the AuthMail struct
	arg := AuthMail{
		Content: AuthMailContent{
			Title:          fmt.Sprintf("%s %s", otpCode, details.Title),
			Name:           user.FirstName,
			Message:        template.HTML(details.Message), // Cast to template.HTML
			Action:         details.Action,
			Link:           fmt.Sprintf("%s/reset-password/%s", config.Domain, otpCode), // Use otpCode
			OTP:            otpCode,                                                     // Use otpCode
			Concern:        details.Concern,
			CompanyName:    config.CompanyName,
			CompanyURL:     config.CompanyURL,
			CompanyAddress: config.CompanyAddress,
			SupportEmail:   config.SupportEmail,
		},
		To: []string{user.Email},
	}

	err = prc.sendAuthEmail(arg)
	if err != nil {
		return fmt.Errorf("failed to send authentication email: %w", err)
	}

	log.Info().Str("type", task.Type()).Bytes("payload", task.Payload()).
		Str("email", user.Email).Msg("processed task")
	return nil
}

// sendAuthEmail is responsible for rendering and sending the email.
func (pr *RedisTaskProcessor) sendAuthEmail(arg AuthMail) error {
	arg.Content.Year = time.Now().Year() // Set the current year

	buff := new(bytes.Buffer)
	// Ensure the template name 'auth.html' matches what you load into pr.template
	err := pr.template.ExecuteTemplate(buff, "auth.html", arg.Content)
	if err != nil {
		log.Error().Err(err).Msg("Failed to execute auth email template")
		return err
	}
	return pr.mailer.SendEmail(mailer.EmailMessage{
		Subject: arg.Content.Title,
		Body:    buff.Bytes(),
		To:      arg.To,
	})
}
