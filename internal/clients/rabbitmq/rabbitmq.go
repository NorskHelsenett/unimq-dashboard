package rabbitmq

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"sync"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type RMQClient struct {
	restClient *rest.RestClient
}

// TODO: Figure out if this is necessary.

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

func NewRMQClient(ctx context.Context, url, username, password string) (*RMQClient, error) {
	restclient, err := rest.NewRestClient(url,
		rest.WithContext(ctx),
		rest.WithUsername(username),
		rest.WithPassword(password),
		rest.WithAuthProvider(httpauthproviders.NewBasicAuthProvider(username, password)),
	)
	if err != nil {
		return nil, err
	}
	return &RMQClient{restClient: restclient}, nil
}

func (r *RMQClient) GetVhosts() ([]string, error) {
	var vhosts []string
	_, err := r.restClient.Get("/vhosts", &vhosts)
	if err != nil {
		return nil, err
	}
	return vhosts, nil
}

func (r *RMQClient) GetVhost(name string) (*models.VhostResponse, error) {

	var vhostData models.VhostResponse
	_, err := r.restClient.Get("/vhosts/"+url.PathEscape(name), &vhostData)
	if err != nil {
		return nil, err
	}

	return &vhostData, nil
}

func (r *RMQClient) GetConnections() ([]models.ConnectionResponse, error) {
	var connections []models.ConnectionResponse
	_, err := r.restClient.Get("/connections", &connections)
	if err != nil {
		return nil, err
	}
	return connections, nil
}

func (r *RMQClient) GetChannels() ([]models.ChannelResponse, error) {
	var channels []models.ChannelResponse
	_, err := r.restClient.Get("/channels", &channels)
	if err != nil {
		return nil, err
	}
	return channels, nil
}

func (r *RMQClient) GetQueues() ([]models.QueueAPIResponse, error) {
	var queues []models.QueueAPIResponse
	_, err := r.restClient.Get("/queues", &queues)
	if err != nil {
		return nil, err
	}
	return queues, nil
}

func (r *RMQClient) GetQueue(vhost string) ([]models.QueueAPIResponse, error) {
	var queues []models.QueueAPIResponse
	_, err := r.restClient.Get("/queues/"+url.PathEscape(vhost), &queues)
	if err != nil {
		return nil, err
	}
	return queues, nil
}

func (r *RMQClient) GetNodes() ([]models.NodeAPIResponse, error) {
	var nodes []models.NodeAPIResponse
	_, err := r.restClient.Get("/nodes", &nodes)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (r *RMQClient) GetMetrics(vhost string) (*models.VhostMetrics, error) {

	vhostObject, err := r.GetVhost(vhost)
	if err != nil {
		return nil, fmt.Errorf("error fetching vhost data: %w", err)
	}

	connections, err := r.GetConnections()
	if err != nil {
		return nil, fmt.Errorf("error fetching connections: %w", err)
	}
	connCount := 0
	for _, c := range connections {
		if c.Vhost == vhost {
			connCount++
		}
	}

	channels, err := r.GetChannels()
	if err != nil {
		return nil, err
	}
	chanCount := 0
	for _, c := range channels {
		if c.Vhost == vhost {
			chanCount++
		}
	}

	queues, err := r.GetQueue(vhost)
	if err != nil {
		return nil, err
	}

	return &models.VhostMetrics{
		Name:        vhost,
		Connections: connCount,
		Channels:    chanCount,
		Queues:      len(queues),
		Unacked:     vhostObject.MessagesUnacknowledged,
	}, nil
}

func (r *RMQClient) GetQueueDetails(vhost string) ([]models.QueueDetail, error) {

	queues, err := r.GetQueue(vhost)
	if err != nil {
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

func (r *RMQClient) GetClusterStats() (*models.ClusterStats, error) {
	nodes, err := r.GetNodes()
	if err != nil {
		return nil, err
	}

	stats := models.NewClusterStats()
	for _, n := range nodes {
		stats.Nodes = append(stats.Nodes, models.NodeStats{
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

	queues, err := r.GetQueues()
	if err != nil {
		return nil, err
	}

	vhostMap := make(map[string]*models.VhostResources)
	for _, q := range queues {
		if _, ok := vhostMap[q.Vhost]; !ok {
			vhostMap[q.Vhost] = &models.VhostResources{Name: q.Vhost}
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
