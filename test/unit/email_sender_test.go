package unit

import (
	"testing"

	config "github.com/bstevary/hexagonal/config"
	"github.com/bstevary/hexagonal/services/mailer"

	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := config.LoadEnv("../../")
	require.NoError(t, err)

	sender := mailer.NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="http://bstevary.com">phcmis Lap</a></p>
	`
	to := []string{"techschool.guru@gmail.com"}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(mailer.EmailMessage{
		Subject:     subject,
		Body:        []byte(content),
		To:          to,
		Attachments: attachFiles,
	})
	require.NoError(t, err)
}
