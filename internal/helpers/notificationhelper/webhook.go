package notificationhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

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
	tctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(tctx, http.MethodPost, url, bytes.NewBuffer(payload))
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
	statuses := make([]WebhookStatus, 0, len(urls))
	for _, url := range urls {
		status := SendWebhook(context.Background(), url, subject, body)
		statuses = append(statuses, *status)
	}

	return statuses
}
