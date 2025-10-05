package mailer

type MailSender interface {
	SendEmail(msg EmailMessage) error
}

type EmailMessage struct {
	Subject     string
	Body        []byte
	To          []string
	Cc          []string
	Bcc         []string
	Attachments []string
}
