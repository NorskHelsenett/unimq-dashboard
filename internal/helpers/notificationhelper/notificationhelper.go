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

type NotifyStatus struct {
	WebhookURLs     []string        `json:"webhook_urls"`
	WebhookStatuses []WebhookStatus `json:"webhook_statuses"`
	EmailRecipients []string        `json:"email_recipients"`
	EmailStatuses   []EmailStatus   `json:"email_statuses"`
}

func NewNotifyStatus(webhookURLs []string, emailRecipients []string) *NotifyStatus {
	return &NotifyStatus{
		WebhookURLs:     webhookURLs,
		WebhookStatuses: make([]WebhookStatus, len(webhookURLs)),
		EmailRecipients: emailRecipients,
		EmailStatuses:   make([]EmailStatus, len(emailRecipients)),
	}
}

func (nh *NotifyStatus) HasErrors() bool {
	for _, ws := range nh.WebhookStatuses {
		if !ws.OK {
			return true
		}
	}
	for _, es := range nh.EmailStatuses {
		if !es.OK {
			return true
		}
	}
	return false
}

func (nh *NotifyStatus) IsTotalFailure() bool {
	for _, ws := range nh.WebhookStatuses {
		if ws.OK {
			return false
		}
	}
	for _, es := range nh.EmailStatuses {
		if es.OK {
			return false
		}
	}
	return true
}

func (nh *NotifyStatus) IsPartialFailure() bool {
	return nh.HasErrors() && !nh.IsTotalFailure()
}

func (nh *NotifyStatus) IsTotalSuccess() bool {
	for _, ws := range nh.WebhookStatuses {
		if !ws.OK {
			return false
		}
	}
	for _, es := range nh.EmailStatuses {
		if !es.OK {
			return false
		}
	}
	return true
}

func (nh *NotifyStatus) FailedDestinations() []string {
	failed := make([]string, 0)
	for _, ws := range nh.WebhookStatuses {
		if !ws.OK {
			failed = append(failed, ws.URL)
		}
	}
	for _, es := range nh.EmailStatuses {
		if !es.OK {
			failed = append(failed, es.Recipient)
		}
	}
	return failed
}

type WebhookStatus struct {
	URL   string `json:"url"`
	OK    bool   `json:"ok"`
	Error error  `json:"error"`
}

func SendWebhook(ctx context.Context, url string, subject, body string) *WebhookStatus {
	status := &WebhookStatus{
		URL:   url,
		OK:    false,
		Error: nil,
	}
	text := subject + "\n\n" + body
	payload, _ := json.Marshal(map[string]string{"text": text})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		status.Error = fmt.Errorf("failed to create request for webhook %s: %w", url, err)
		slog.ErrorContext(ctx, "failed to create request for webhook", "url", url, "error", err)
		return status
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		status.Error = fmt.Errorf("failed to send request to webhook %s: %w", url, err)
		slog.ErrorContext(ctx, "failed to send request to webhook", "url", url, "error", err)
		return status
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.ErrorContext(ctx, "failed to close response body for webhook", "url", url, "error", err)
		}
	}()

	if resp.StatusCode >= 400 {
		status.Error = fmt.Errorf("webhook returned %d", resp.StatusCode)
		slog.ErrorContext(ctx, "webhook returned error status code", "url", url, "status_code", resp.StatusCode)
		return status
	}
	status.OK = true
	return status
}

func SendWebhooks(urls []string, subject, body string) []WebhookStatus {
	var statuses []WebhookStatus
	for _, url := range urls {
		status := SendWebhook(context.Background(), url, subject, body)
		statuses = append(statuses, *status)
	}

	return statuses
}

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
	var statuses []EmailStatus
	for _, email := range to {
		status := es.SendEmail(email, subject, body, typ)
		statuses = append(statuses, *status)
	}

	return statuses
}
