package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"sync"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/prom"
)

const historySize = 20

var history = struct {
	mu   sync.Mutex
	data map[string][]int
}{data: make(map[string][]int)}

func appendHistory(key string, value int) []int {
	history.mu.Lock()
	defer history.mu.Unlock()
	h := append(history.data[key], value)
	if len(h) > historySize {
		h = h[len(h)-historySize:]
	}
	history.data[key] = h
	return h
}

// TODO: Move to env vars or config file
var DefaultLimits = models.Limits{
	MaxChannels:    1000,
	MaxConnections: 300,
	MaxQueues:      150,
}

// ClusterStats holds cluster-wide memory and disk info plus per-vhost breakdown.
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
	PublishDetails rateDetail `json:"publish_details"`
	DeliverDetails rateDetail `json:"deliver_get_details"`
	RedelivDetails rateDetail `json:"redeliver_details"`
}

type queueAPIResponse struct {
	Name                   string       `json:"name"`
	Vhost                  string       `json:"vhost"`
	Messages               int          `json:"messages"`
	MessagesUnacknowledged int          `json:"messages_unacknowledged"`
	Consumers              int          `json:"consumers"`
	MessageBytes           int64        `json:"message_bytes"`
	MessageBytesPersistent int64        `json:"message_bytes_persistent"`
	MessageStats           messageStats `json:"message_stats"`
}

type nodeAPIResponse struct {
	Name          string `json:"name"`
	MemUsed       int64  `json:"mem_used"`
	MemLimit      int64  `json:"mem_limit"`
	DiskFree      int64  `json:"disk_free"`
	DiskFreeLimit int64  `json:"disk_free_limit"`
}

// TODO: Seperate the clients
type RestClient struct {
	baseURL    string
	username   string
	password   string
	PromClient *prom.PromClient
}

func NewRestClient(baseURL, username, password string, promURL, promAPIVersion string, promPort int) *RestClient {
	return &RestClient{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		PromClient: prom.NewPromClient(promURL, promAPIVersion, promPort),
	}
}

// TODO: Is Get request.
// TODO: Make a generic rest client (Borrow from ror-ms-backup)
func (r *RestClient) fetch(path string, v any) error {
	req, err := http.NewRequest("GET", r.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(r.username, r.password)

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

func (r *RestClient) GetVhosts() ([]string, error) {
	var vhosts []vhostResponse
	if err := r.fetch("/vhosts", &vhosts); err != nil {
		return nil, err
	}
	names := make([]string, len(vhosts))
	for i, v := range vhosts {
		names[i] = v.Name
	}
	return names, nil
}

func (r *RestClient) GetMetrics(vhost string) (*models.VhostMetrics, error) {
	encoded := url.PathEscape(vhost)

	var vhostData vhostResponse
	if err := r.fetch("/vhosts/"+encoded, &vhostData); err != nil {
		return nil, err
	}

	var connections []connectionResponse
	if err := r.fetch("/connections", &connections); err != nil {
		return nil, err
	}
	connCount := 0
	for _, c := range connections {
		if c.Vhost == vhost {
			connCount++
		}
	}

	var channels []channelResponse
	if err := r.fetch("/channels", &channels); err != nil {
		return nil, err
	}
	chanCount := 0
	for _, c := range channels {
		if c.Vhost == vhost {
			chanCount++
		}
	}

	var queues []queueAPIResponse
	if err := r.fetch("/queues/"+encoded, &queues); err != nil {
		return nil, err
	}

	return &models.VhostMetrics{
		Name:        vhost,
		Connections: connCount,
		Channels:    chanCount,
		Queues:      len(queues),
		Unacked:     vhostData.MessagesUnacknowledged,
	}, nil
}

func (r *RestClient) GetQueueDetails(vhost string) ([]models.QueueDetail, error) {
	encoded := url.PathEscape(vhost)

	var queues []queueAPIResponse
	if err := r.fetch("/queues/"+encoded, &queues); err != nil {
		return nil, err
	}

	details := make([]models.QueueDetail, len(queues))
	for i, q := range queues {
		key := vhost + "/" + q.Name
		details[i] = models.QueueDetail{
			Name:         q.Name,
			Messages:     q.Messages,
			MessageBytes: q.MessageBytes,
			History:      appendHistory(key, q.Messages),
			Consumers:    q.Consumers,
			PublishRate:  q.MessageStats.PublishDetails.Rate,
			DeliverRate:  q.MessageStats.DeliverDetails.Rate,
			RedelivRate:  q.MessageStats.RedelivDetails.Rate,
			Unacked:      q.MessagesUnacknowledged,
		}
	}
	return details, nil
}

func (r *RestClient) GetClusterStats() (*ClusterStats, error) {
	var nodes []nodeAPIResponse
	if err := r.fetch("/nodes", &nodes); err != nil {
		return nil, err
	}

	stats := &ClusterStats{}
	for _, n := range nodes {
		stats.Nodes = append(stats.Nodes, NodeStats{
			Name:          n.Name,
			MemUsed:       n.MemUsed,
			MemLimit:      n.MemLimit,
			DiskFree:      n.DiskFree,
			DiskFreeLimit: n.DiskFreeLimit,
		})
		stats.TotalMemUsed += n.MemUsed
		stats.TotalMemLimit += n.MemLimit
		stats.TotalDiskFree += n.DiskFree
		if stats.MinDiskLimit == 0 || n.DiskFreeLimit < stats.MinDiskLimit {
			stats.MinDiskLimit = n.DiskFreeLimit
		}
	}

	// Aggregate per-vhost message bytes from all queues
	var allQueues []queueAPIResponse
	if err := r.fetch("/queues", &allQueues); err != nil {
		return nil, err
	}

	vhostMap := make(map[string]*VhostResources)
	for _, q := range allQueues {
		if _, ok := vhostMap[q.Vhost]; !ok {
			vhostMap[q.Vhost] = &VhostResources{Name: q.Vhost}
		}
		vhostMap[q.Vhost].MessageBytes += q.MessageBytes
		vhostMap[q.Vhost].DiskBytes += q.MessageBytesPersistent
	}

	for _, v := range vhostMap {
		stats.VhostResources = append(stats.VhostResources, *v)
	}
	sort.Slice(stats.VhostResources, func(i, j int) bool {
		return stats.VhostResources[i].MessageBytes > stats.VhostResources[j].MessageBytes
	})

	return stats, nil
}
