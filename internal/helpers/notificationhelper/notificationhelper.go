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
		err = resp.Body.Close()
		if err != nil {
			lastErr = err
		}
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("webhook returned %d", resp.StatusCode)
		}
	}
	return lastErr
}
