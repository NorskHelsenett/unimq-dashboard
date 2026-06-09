package models

type VhostMetrics struct {
	Name        string `json:"name"`
	Connections int    `json:"connections"`
	Channels    int    `json:"channels"`
	Queues      int    `json:"queues"`
	Unacked     int    `json:"unacked"`
}

type QueueDetail struct {
	Name         string  `json:"name"`
	Messages     int     `json:"messages"`
	MessageBytes int64   `json:"message_bytes"`
	History      []int   `json:"history"`
	Consumers    int     `json:"consumers"`
	PublishRate  float64 `json:"publish_rate"`
	DeliverRate  float64 `json:"deliver_rate"`
	RedelivRate  float64 `json:"redeliver_rate"`
	Unacked      int     `json:"messages_unacknowledged"`
}

type Limits struct {
	MaxChannels    int
	MaxConnections int
	MaxQueues      int
}
