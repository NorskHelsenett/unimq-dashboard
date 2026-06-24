package models

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

type Metadata struct {
	Description      string   `json:"description"`
	Tags             []string `json:"tags"`
	DefaultQueueType string   `json:"default_queue_type"`
}

type MessageRate struct {
	Rate float64 `json:"rate"`
}

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

type NodeStats struct {
	Name          string `json:"name"`
	MemUsed       int64  `json:"mem_used"`
	MemLimit      int64  `json:"mem_limit"`
	DiskFree      int64  `json:"disk_free"`
	DiskFreeLimit int64  `json:"disk_free_limit"`
}

type VhostResources struct {
	Name         string `json:"name"`
	MessageBytes int64  `json:"message_bytes"`
	DiskBytes    int64  `json:"disk_bytes"`
}

type ClusterStats struct {
	Nodes          []NodeStats      `json:"nodes"`
	TotalMemUsed   int64            `json:"total_mem_used"`
	TotalMemLimit  int64            `json:"total_mem_limit"`
	TotalDiskFree  int64            `json:"total_disk_free"`
	MinDiskLimit   int64            `json:"min_disk_limit"`
	VhostResources []VhostResources `json:"vhost_resources"`
}

func NewClusterStats() *ClusterStats {
	return &ClusterStats{
		Nodes:          []NodeStats{},
		TotalMemUsed:   0,
		TotalMemLimit:  0,
		TotalDiskFree:  0,
		MinDiskLimit:   0,
		VhostResources: []VhostResources{},
	}
}

type ConnectionResponse struct {
	Vhost string `json:"vhost"`
}

type ChannelResponse struct {
	Vhost string `json:"vhost"`
}

type RateDetail struct {
	Rate float64 `json:"rate"`
}

type MessageStats struct {
	PublishDetails RateDetail `json:"publish_details"`
	DeliverDetails RateDetail `json:"deliver_get_details"`
	RedelivDetails RateDetail `json:"redeliver_details"`
}

type QueueAPIResponse struct {
	Name                   string       `json:"name"`
	Vhost                  string       `json:"vhost"`
	Messages               int          `json:"messages"`
	MessagesUnacknowledged int          `json:"messages_unacknowledged"`
	Consumers              int          `json:"consumers"`
	MessageBytes           int64        `json:"message_bytes"`
	MessageBytesPersistent int64        `json:"message_bytes_persistent"`
	MessageStats           MessageStats `json:"message_stats"`
}
