package mailer

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Constants for Gmail SMTP
const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

// GmailSender implements the MailSender interface using Gmail SMTP.
type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) MailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

// SendEmail converts the public EmailMessage into an internal email and sends it.
func (s *GmailSender) SendEmail(msg EmailMessage) error {
	// 1. Convert public message to internal MIME representation
	e := s.newMIMEEmail(msg)

	// 2. Attach files
	for _, f := range msg.Attachments {
		if err := e.attachFile(f); err != nil {
			return fmt.Errorf("failed to attach file %s: %w", f, err)
		}
	}

	// 3. Prepare Authentication
	auth := smtp.PlainAuth("", s.fromEmailAddress, s.fromEmailPassword, smtpAuthAddress)

	// 4. Send the email using the internal logic
	return e.sendMail(smtpServerAddress, auth)
}

// ====================================================================
// INTERNAL IMPLEMENTATION (UNEXPORTED TYPES AND METHODS)
// ====================================================================

// attachment is an unexported struct representing an email attachment.
type attachment struct {
	Filename    string
	ContentType string
	Header      textproto.MIMEHeader
	Content     []byte
}

type mimeEmail struct {
	From        string
	To          []string
	Bcc         []string
	Cc          []string
	Subject     string
	HTML        []byte
	Attachments []*attachment
}

func (s *GmailSender) newMIMEEmail(msg EmailMessage) *mimeEmail {
	return &mimeEmail{
		From:    fmt.Sprintf("%s <%s>", s.name, s.fromEmailAddress),
		Subject: msg.Subject,
		HTML:    msg.Body,
		To:      msg.To,
		Cc:      msg.Cc,
		Bcc:     msg.Bcc,
	}
}

// attachFile opens a file, reads its content, determines its MIME type, and attaches it.
func (e *mimeEmail) attachFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open attachment file: %w", err)
	}
	defer f.Close()

	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, f); err != nil {
		return fmt.Errorf("failed to read attachment content: %w", err)
	}

	ct := mime.TypeByExtension(filepath.Ext(filename))
	if ct == "" {
		ct = "application/octet-stream"
	}

	basename := filepath.Base(filename)

	at := &attachment{
		Filename:    basename,
		ContentType: ct,
		Header:      textproto.MIMEHeader{},
		Content:     buffer.Bytes(),
	}
	e.Attachments = append(e.Attachments, at)
	return nil
}

// buildHeaders constructs the necessary email headers.
func (e *mimeEmail) buildHeaders() (textproto.MIMEHeader, error) {
	if e.From == "" || len(e.To) == 0 {
		return nil, errors.New("must specify at least one From address and one To address")
	}

	res := make(textproto.MIMEHeader)
	res.Set("From", e.From)
	res.Set("To", strings.Join(e.To, ", "))

	if len(e.Cc) > 0 {
		res.Set("Cc", strings.Join(e.Cc, ", "))
	}
	if len(e.Bcc) > 0 {
		// Bcc is intentionally omitted from headers for privacy
	}
	res.Set("Subject", e.Subject)
	res.Set("MIME-Version", "1.0")
	res.Set("Date", time.Now().Format(time.RFC1123Z))

	return res, nil
}

// bytes converts the internal email object to a []byte MIME representation and returns the SMTP recipients list.
func (e *mimeEmail) bytes() ([]byte, []string, error) {
	// 1. Get headers
	headers, err := e.buildHeaders()
	if err != nil {
		return nil, nil, err
	}

	// 2. Prepare recipients list for SMTP envelope (includes Bcc)
	recipients := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))

	// Helper to parse and collect addresses
	collectRecipients := func(addrs []string) error {
		for _, addr := range addrs {
			parsed, err := mail.ParseAddress(addr)
			if err != nil {
				return err
			}
			recipients = append(recipients, parsed.Address)
		}
		return nil
	}

	if err := collectRecipients(e.To); err != nil {
		return nil, nil, err
	}
	if err := collectRecipients(e.Cc); err != nil {
		return nil, nil, err
	}
	if err := collectRecipients(e.Bcc); err != nil {
		return nil, nil, err
	}

	// 3. Start buffer and multipart writer
	buf := new(bytes.Buffer)

	// Determine the main Content-Type: multipart/mixed (for body + attachments)
	w := multipart.NewWriter(buf)
	headers.Set("Content-Type", "multipart/mixed;\r\n boundary="+w.Boundary())

	// Write headers to buffer
	for k, v := range headers {
		buf.WriteString(k + ": " + strings.Join(v, ", ") + "\r\n")
	}
	buf.WriteString("\r\n") // End of headers, start of body

	// 4. Write HTML Body Part (as text/html, quoted-printable)
	htmlHeader := textproto.MIMEHeader{
		"Content-Type":              {"text/html; charset=UTF-8"},
		"Content-Transfer-Encoding": {"quoted-printable"},
	}
	bodyPart, err := w.CreatePart(htmlHeader)
	if err != nil {
		return nil, nil, err
	}

	qp := quotedprintable.NewWriter(bodyPart)
	if _, err := qp.Write(e.HTML); err != nil {
		return nil, nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, nil, err
	}

	// 5. Write Attachment Parts (as application/octet-stream, base64)
	for _, a := range e.Attachments {
		// Define attachment headers
		attHeader := textproto.MIMEHeader{
			"Content-Type":              {fmt.Sprintf("%s;\r\n name=\"%s\"", a.ContentType, a.Filename)},
			"Content-Transfer-Encoding": {"base64"},
			"Content-Disposition":       {fmt.Sprintf("attachment;\r\n filename=\"%s\"", a.Filename)},
		}

		// Create the attachment part
		ap, err := w.CreatePart(attHeader)
		if err != nil {
			return nil, nil, err
		}

		// Write the content base64 encoded
		encoder := base64.NewEncoder(base64.StdEncoding, ap)
		if _, err := encoder.Write(a.Content); err != nil {
			return nil, nil, err
		}
		if err := encoder.Close(); err != nil {
			return nil, nil, err
		}
	}

	// 6. Close the main multipart writer
	if err := w.Close(); err != nil {
		return nil, nil, err
	}

	return buf.Bytes(), recipients, nil
}

// sendMail sends the email using the provided SMTP address and authentication.
func (e *mimeEmail) sendMail(addr string, a smtp.Auth) error {
	raw, recipients, err := e.bytes()
	if err != nil {
		return err
	}

	// Determine the sender address for the SMTP envelope
	senderAddr, err := mail.ParseAddress(e.From)
	if err != nil {
		return err
	}
	sender := senderAddr.Address

	return smtp.SendMail(addr, a, sender, recipients, raw)
}
