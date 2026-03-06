package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	baseURL  = "http://localhost:15672/api"
	username = "guest"
	password = "guest"
)

type VhostMetrics struct {
	Name        string `json:"name"`
	Connections int    `json:"connections"`
	Channels    int    `json:"channels"`
	Queues      int    `json:"queues"`
	Unacked     int    `json:"unacked"`
}

type Limits struct {
	MaxConnections int
	MaxQueues      int
	MaxMessages    int
	MaxQueueSize   string
}

var DefaultLimits = Limits{
	MaxConnections: 300,
	MaxQueues:      150,
	MaxMessages:    10000,
	MaxQueueSize:   "10 GiB",
}

type vhostResponse struct {
	Name  string `json:"name"`
	Stats struct {
		MessagesUnacknowledged int `json:"messages_unacknowledged"`
	} `json:"messages_details"`
	MessagesUnacknowledged int `json:"messages_unacknowledged"`
}

type connectionResponse struct {
	Vhost string `json:"vhost"`
}

type channelResponse struct {
	Vhost string `json:"vhost"`
}

type queueResponse struct {
	Vhost string `json:"vhost"`
}

func fetch(path string, v any) error {
	req, err := http.NewRequest("GET", baseURL+path, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach RabbitMQ Management API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

func GetVhosts() ([]string, error) {
	var vhosts []vhostResponse
	if err := fetch("/vhosts", &vhosts); err != nil {
		return nil, err
	}
	names := make([]string, len(vhosts))
	for i, v := range vhosts {
		names[i] = v.Name
	}
	return names, nil
}

func GetMetrics(vhost string) (*VhostMetrics, error) {
	encoded := url.PathEscape(vhost)

	var vhostData vhostResponse
	if err := fetch("/vhosts/"+encoded, &vhostData); err != nil {
		return nil, err
	}

	var connections []connectionResponse
	if err := fetch("/connections", &connections); err != nil {
		return nil, err
	}
	connCount := 0
	for _, c := range connections {
		if c.Vhost == vhost {
			connCount++
		}
	}

	var channels []channelResponse
	if err := fetch("/channels", &channels); err != nil {
		return nil, err
	}
	chanCount := 0
	for _, c := range channels {
		if c.Vhost == vhost {
			chanCount++
		}
	}

	var queues []queueResponse
	if err := fetch("/queues/"+encoded, &queues); err != nil {
		return nil, err
	}

	return &VhostMetrics{
		Name:        vhost,
		Connections: connCount,
		Channels:    chanCount,
		Queues:      len(queues),
		Unacked:     vhostData.MessagesUnacknowledged,
	}, nil
}
