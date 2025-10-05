package mailer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/resend/resend-go/v2"
)

type resendSender struct {
	client *resend.Client
	from   string
}

func NewResendSender(from, apiKey string) MailSender {
	client := resend.NewClient(apiKey)
	return &resendSender{
		client: client,
		from:   from,
	}
}

func (s *resendSender) SendEmail(msg EmailMessage) error {
	attachments, err := buildResendSDKAttachments(msg.Attachments)
	if err != nil {
		return fmt.Errorf("failed to prepare attachments: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:        s.from,
		To:          msg.To,
		Cc:          msg.Cc,
		Bcc:         msg.Bcc,
		Subject:     msg.Subject,
		Html:        string(msg.Body),
		Attachments: attachments,
	}

	_, err = s.client.Emails.SendWithContext(context.Background(), params)
	if err != nil {
		return fmt.Errorf("resend send failed: %w", err)
	}

	return nil
}

func buildResendSDKAttachments(paths []string) ([]*resend.Attachment, error) {
	var attachments []*resend.Attachment
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", path, err)
		}
		attachments = append(attachments, &resend.Attachment{
			Content:  data,
			Filename: filepath.Base(path),
		})
	}
	return attachments, nil
}
