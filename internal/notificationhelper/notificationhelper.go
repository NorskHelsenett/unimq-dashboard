package notificationhelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func SendWebhooks(urls []string, subject, body string) error {
	text := subject + "\n\n" + body
	payload, _ := json.Marshal(map[string]string{"text": text})
	var lastErr error
	for _, u := range urls {
		resp, err := http.Post(u, "application/json", bytes.NewReader(payload))
		if err != nil {
			lastErr = err
			continue
		}
		resp.Body.Close()
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("webhook returnerte HTTP %d", resp.StatusCode)
		}
	}
	return lastErr
}
