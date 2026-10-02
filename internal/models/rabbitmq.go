package models

import "encoding/json"

type Vhost struct {
	Messages                      int               `json:"messages"`
	Name                          string            `json:"name"`
	Description                   string            `json:"description"`
	Metadata                      Metadata          `json:"metadata"`
	Tags                          []string          `json:"tags"`
	DefaultQueueType              string            `json:"default_queue_type"`
	MessagesReady                 int               `json:"messages_ready"`
	MessagesUnacknowledged        int               `json:"messages_unacknowledged"`
	ProtectedFromDeletion         bool              `json:"protected_from_deletion"`
	Tracing                       bool              `json:"tracing"`
	ClusterState                  map[string]string `json:"cluster_state"`
	MessagesDetails               MessageRate       `json:"messages_details"`
	MessagesUnacknowledgedDetails MessageRate       `json:"messages_unacknowledged_details"`
	MessagesReadyDetails          MessageRate       `json:"messages_ready_details"`
}

type VhostPost struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	DefaultQueueType string   `json:"default_queue_type"`
}

type Metadata struct {
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	DefaultQueueType string   `json:"default_queue_type"`
}

type MessageRate struct {
	Rate float64 `json:"rate"`
}

type VhostMetrics struct {
	Name            string `json:"name"`
	Connections     int    `json:"connections"`
	Channels        int    `json:"channels"`
	Queues          int    `json:"queues"`
	UnackedMessages int    `json:"unacked_messages"`
	ReadyMessages   int    `json:"ready_messages"`
}

type RMQVhostUsage struct {
	Name         string `json:"name"`
	MessageBytes int64  `json:"message_bytes"`
	DiskBytes    int64  `json:"disk_bytes"`
}

type RMQNode struct {
	Name          string `json:"name"`
	MemUsed       int64  `json:"mem_used"`
	MemLimit      int64  `json:"mem_limit"`
	DiskFree      int64  `json:"disk_free"`
	DiskFreeLimit int64  `json:"disk_free_limit"`
}

func NewRMQNode(name string, memUsed, memLimit, diskFree, diskFreeLimit int64) *RMQNode {
	return &RMQNode{
		Name:          name,
		MemUsed:       memUsed,
		MemLimit:      memLimit,
		DiskFree:      diskFree,
		DiskFreeLimit: diskFreeLimit,
	}
}

type RMQVhostLimits struct {
	Vhost          string `json:"vhost"`
	MaxConnections int    `json:"max_connections"`
	MaxQueues      int    `json:"max_queues"`
}

func NewRMQVhostLimits(vhost string) *RMQVhostLimits {
	return &RMQVhostLimits{
		Vhost:          vhost,
		MaxConnections: 0,
		MaxQueues:      0,
	}
}

func (l *RMQVhostLimits) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Vhost  string `json:"vhost"`
		Values struct {
			MaxConnections *int `json:"max_connections"`
			MaxQueues      *int `json:"max_queues"`
		} `json:"values"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Values.MaxConnections != nil {
		l.MaxConnections = *aux.Values.MaxConnections
	} else {
		l.MaxConnections = 0
	}

	if aux.Values.MaxQueues != nil {
		l.MaxQueues = *aux.Values.MaxQueues
	} else {
		l.MaxQueues = 0
	}

	return nil
}

type RMQConnection struct {
	Vhost string `json:"vhost"`
}

type RMQChannel struct {
	Vhost string `json:"vhost"`
}

type RMQQueue struct {
	Name                   string  `json:"name"`
	Vhost                  string  `json:"vhost"`
	Messages               int     `json:"messages"`
	MessagesUnacknowledged int     `json:"messages_unacknowledged"`
	Consumers              int     `json:"consumers"`
	MessageBytes           int64   `json:"message_bytes"`
	MessageBytesPersistent int64   `json:"message_bytes_persistent"`
	History                []int   `json:"history"`
	PublishRate            float64 `json:"publish_rate"`
	DeliverRate            float64 `json:"deliver_rate"`
	RedliverRate           float64 `json:"redeliver_rate"`
}

func (q *RMQQueue) UnmarshalJSON(data []byte) error {
	type Alias RMQQueue
	aux := &struct {
		MessageStats struct {
			PublishDetails struct {
				Rate float64 `json:"rate"`
			} `json:"publish_details"`
			DeliverDetails struct {
				Rate float64 `json:"rate"`
			} `json:"deliver_details"`
			RedelivDetails struct {
				Rate float64 `json:"rate"`
			} `json:"redeliver_details"`
		} `json:"message_stats"`
		*Alias
	}{
		Alias: (*Alias)(q),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	q.PublishRate = aux.MessageStats.PublishDetails.Rate
	q.DeliverRate = aux.MessageStats.DeliverDetails.Rate
	q.RedliverRate = aux.MessageStats.RedelivDetails.Rate

	return nil
}

type QueuePost struct {
	Name       string `json:"name"`
	AutoDelete bool   `json:"auto_delete"`
	Durable    bool   `json:"durable"`
	Vhost      string `json:"vhost"`
	Arguments  any    `json:"arguments"`
}
