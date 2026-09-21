package notificationhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/wneessen/go-mail"
)

func SendWebhooks(urls []string, subject, body string) error {
	text := subject + "\n\n" + body
	payload, _ := json.Marshal(map[string]string{"text": text})
	var lastErr error
	for _, u := range urls {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBuffer(payload))
		if err != nil {
			slog.ErrorContext(ctx, "failed to create request for webhook", "url", u, "error", err)
			lastErr = fmt.Errorf("failed to create request for webhook %s: %w", u, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			slog.ErrorContext(ctx, "failed to send request to webhook", "url", u, "error", err)
			lastErr = err
			continue
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				slog.ErrorContext(ctx, "failed to close response body for webhook", "url", u, "error", err)
				lastErr = err
			}
		}()

		if resp.StatusCode >= 400 {
			slog.ErrorContext(ctx, "webhook returned error status code", "url", u, "status_code", resp.StatusCode)
			lastErr = fmt.Errorf("webhook returned %d", resp.StatusCode)
		}
	}
	return lastErr
}

var (
	ErrEmailNotConfigured = fmt.Errorf("SMTP server is not configured")
	ErrEmailSendFailed    = fmt.Errorf("failed to send email")
)

type emailSenderError struct {
	destinations []emailDestinationError
}

type emailDestinationError struct {
	destination string
	err         error
}

func (e *emailSenderError) Error() string {
	var buffer bytes.Buffer
	buffer.WriteString("failed to send email to the following destinations:\n")
	for _, destErr := range e.destinations {
		_, err := fmt.Fprintf(&buffer, "- %s: %v\n", destErr.destination, destErr.err)
		if err != nil {
			slog.Error("failed to write to buffer", "error", err)
		}

	}
	return buffer.String()
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

func (es *EmailSender) SendEmail(to, subject, body string, typ mail.ContentType) error {
	return sendEmail(es.Config, to, subject, body, typ)
}

func (es *EmailSender) SendEmails(ctx context.Context, to []string, subject, body string, typ mail.ContentType) error {
	sendErr := emailSenderError{
		destinations: make([]emailDestinationError, 0),
	}
	for _, email := range to {
		err := es.SendEmail(email, subject, body, typ)
		if err != nil {
			if errors.Is(err, ErrEmailNotConfigured) {
				return ErrEmailNotConfigured
			}
			sendErr.destinations = append(sendErr.destinations, emailDestinationError{
				destination: email,
				err:         err,
			})
			continue
		}

	}

	if len(sendErr.destinations) > 0 {
		return &sendErr
	}

	return nil
}

func sendEmail(config *config.EmailConfig, to, subject, body string, typ mail.ContentType) error {

	if config == nil {
		return ErrEmailNotConfigured
	}

	if config.EmailFromAddress == "" {
		return ErrEmailNotConfigured
	}
	message := mail.NewMsg()
	if err := message.From(config.EmailFromAddress); err != nil {
		return fmt.Errorf("failed to set From address: %w", err)
	}
	if err := message.To(to); err != nil {
		return fmt.Errorf("failed to set To address: %w", err)
	}

	message.Subject(subject)
	message.SetBodyString(typ, body)

	client, err := mail.NewClient(config.EmailSMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := client.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}

	return nil
}
