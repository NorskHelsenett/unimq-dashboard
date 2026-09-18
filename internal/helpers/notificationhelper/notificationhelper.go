package notificationhelper

import (
	"bytes"
	"context"
	"encoding/json"
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

var ErrEmailNotConfigured = fmt.Errorf("SMTP server is not configured")

func SendEmail(config *config.EmailConfig, to, subject, body string, typ mail.ContentType) error {

	if config == nil {
		return ErrEmailNotConfigured
	}

	if config.EmailFromAddress == "" {
		return fmt.Errorf("SMTP server is not configured")
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
