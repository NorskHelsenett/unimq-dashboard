package notificationhelper

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/wneessen/go-mail"
)

var (
	ErrEmailNotConfigured = fmt.Errorf("SMTP server is not configured")
	ErrEmailSendFailed    = fmt.Errorf("failed to send email")
)

type EmailStatus struct {
	Recipient string `json:"recipient"`
	OK        bool   `json:"ok"`
	Error     error  `json:"error"`
}

func sendEmail(config *config.EmailConfig, to, subject, body string, typ mail.ContentType) *EmailStatus {
	status := &EmailStatus{
		Recipient: to,
		OK:        false,
		Error:     nil,
	}

	if config == nil {
		status.Error = ErrEmailNotConfigured
		return status
	}

	if config.EmailFromAddress == "" {
		status.Error = ErrEmailNotConfigured
		return status
	}
	message := mail.NewMsg()
	if err := message.From(config.EmailFromAddress); err != nil {
		status.Error = fmt.Errorf("failed to set From address: %w", err)
		return status
	}
	if err := message.To(to); err != nil {
		status.Error = fmt.Errorf("failed to set To address: %w", err)
		return status
	}

	message.Subject(subject)
	message.SetBodyString(typ, body)

	client, err := mail.NewClient(config.EmailSMTPHost)
	if err != nil {
		status.Error = fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := client.DialAndSend(message); err != nil {
		status.Error = fmt.Errorf("failed to send mail: %w", err)
		return status
	}

	status.OK = true
	return status
}

type EmailSender struct {
	Config *config.EmailConfig
}

var EmailSenderInstance *EmailSender

func InitEmailSender(config *config.EmailConfig) {
	EmailSenderInstance = &EmailSender{
		Config: config,
	}
}

func (es *EmailSender) SendEmail(to, subject, body string, typ mail.ContentType) *EmailStatus {
	return sendEmail(es.Config, to, subject, body, typ)
}

func (es *EmailSender) SendEmails(ctx context.Context, to []string, subject, body string, typ mail.ContentType) []EmailStatus {
	statuses := make([]EmailStatus, 0, len(to))
	for _, email := range to {
		status := es.SendEmail(email, subject, body, typ)
		statuses = append(statuses, *status)
	}

	return statuses
}
