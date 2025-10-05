package mailer

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const brevoAPIURL = "https://api.brevo.com/v3/smtp/email"

type BrevoSender struct {
	name   string
	from   string
	apiKey string
	client *http.Client
}

func NewBrevoSender(name, from, apiKey string) MailSender {
	return &BrevoSender{
		name:   name,
		from:   from,
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (s *BrevoSender) SendEmail(msg EmailMessage) error {
	payload := map[string]any{
		"sender": map[string]string{
			"name":  s.name,
			"email": s.from,
		},
		"to":          toEmails(msg.To),
		"cc":          toEmails(msg.Cc),
		"bcc":         toEmails(msg.Bcc),
		"subject":     msg.Subject,
		"htmlContent": string(msg.Body),
	}

	if len(msg.Attachments) > 0 {
		attachments, err := buildAttachments(msg.Attachments)
		if err != nil {
			return fmt.Errorf("failed to build attachments: %w", err)
		}
		payload["attachment"] = attachments
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", brevoAPIURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("brevo error: %s", body)
	}

	return nil
}

func toEmails(addresses []string) []map[string]string {
	var list []map[string]string
	for _, addr := range addresses {
		list = append(list, map[string]string{"email": addr})
	}
	return list
}

func buildAttachments(paths []string) ([]map[string]string, error) {
	var attachments []map[string]string
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("could not read file %s: %w", path, err)
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		attachments = append(attachments, map[string]string{
			"name":    filepath.Base(path),
			"content": encoded,
		})
	}
	return attachments, nil
}
