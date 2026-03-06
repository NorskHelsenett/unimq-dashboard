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

type QueueDetail struct {
	Name        string  `json:"name"`
	Messages    int     `json:"messages"`
	Consumers   int     `json:"consumers"`
	PublishRate float64 `json:"publish_rate"`
	DeliverRate float64 `json:"deliver_rate"`
	RedelivRate float64 `json:"redeliver_rate"`
	Unacked     int     `json:"messages_unacknowledged"`
}

type Limits struct {
	MaxConnections int
	MaxQueues      int
}

var DefaultLimits = Limits{
	MaxConnections: 300,
	MaxQueues:      150,
}

type vhostResponse struct {
	Name                   string `json:"name"`
	MessagesUnacknowledged int    `json:"messages_unacknowledged"`
}

type connectionResponse struct {
	Vhost string `json:"vhost"`
}

type channelResponse struct {
	Vhost string `json:"vhost"`
}

type rateDetail struct {
	Rate float64 `json:"rate"`
}

type messageStats struct {
	PublishDetails  rateDetail `json:"publish_details"`
	DeliverDetails  rateDetail `json:"deliver_get_details"`
	RedelivDetails  rateDetail `json:"redeliver_details"`
}

type queueAPIResponse struct {
	Name                   string       `json:"name"`
	Vhost                  string       `json:"vhost"`
	Messages               int          `json:"messages"`
	MessagesUnacknowledged int          `json:"messages_unacknowledged"`
	Consumers              int          `json:"consumers"`
	MessageStats           messageStats `json:"message_stats"`
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

	var queues []queueAPIResponse
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

func GetQueueDetails(vhost string) ([]QueueDetail, error) {
	encoded := url.PathEscape(vhost)

	var queues []queueAPIResponse
	if err := fetch("/queues/"+encoded, &queues); err != nil {
		return nil, err
	}

	details := make([]QueueDetail, len(queues))
	for i, q := range queues {
		details[i] = QueueDetail{
			Name:        q.Name,
			Messages:    q.Messages,
			Consumers:   q.Consumers,
			PublishRate: q.MessageStats.PublishDetails.Rate,
			DeliverRate: q.MessageStats.DeliverDetails.Rate,
			RedelivRate: q.MessageStats.RedelivDetails.Rate,
			Unacked:     q.MessagesUnacknowledged,
		}
	}
	return details, nil
}
