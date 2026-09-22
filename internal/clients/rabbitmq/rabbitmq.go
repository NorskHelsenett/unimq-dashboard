package rabbitmq

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type RMQClient struct {
	restClient *rest.RestClient
	Limits     *models.Limits
}

// TODO: Figure out if the history is necessary.
// If it is, the history should be stored in the database, not in memory, as it will be lost on restart.
// The history length should also be moved to the config.

const historySize = 20

var history = struct {
	mu   sync.Mutex
	data map[string][]int
}{data: make(map[string][]int)}

func appendHistory(key string, value int) []int {
	history.mu.Lock()
	defer history.mu.Unlock()
	// nolint:gocritic // leaving this for now, as use is unclear.
	h := append(history.data[key], value)
	if len(h) > historySize {
		h = h[len(h)-historySize:]
	}
	history.data[key] = h
	return h
}

type (
	rmqClientConfig struct {
		Host     string
		Port     int
		Username string
		Password string
		Ctx      context.Context
	}

	rmqClientOptions func(*rmqClientConfig)
)

func newRMQClientConfig() *rmqClientConfig {
	return &rmqClientConfig{
		Host:     "localhost",
		Port:     15672,
		Username: "",
		Password: "",
		Ctx:      context.Background(),
	}
}

func WithRMQHost(host string) rmqClientOptions {
	return func(rc *rmqClientConfig) {
		rc.Host = host
	}
}

func WithRMQPort(port int) rmqClientOptions {
	return func(rc *rmqClientConfig) {
		rc.Port = port
	}
}

func WithRMQUsername(username string) rmqClientOptions {
	return func(rc *rmqClientConfig) {
		rc.Username = username
	}
}

func WithRMQPassword(password string) rmqClientOptions {
	return func(rc *rmqClientConfig) {
		rc.Password = password
	}
}

func WithRMQContext(ctx context.Context) rmqClientOptions {
	return func(rc *rmqClientConfig) {
		rc.Ctx = ctx
	}
}

func NewRMQClient(opts ...rmqClientOptions) (*RMQClient, error) {
	config := newRMQClientConfig()
	for _, opt := range opts {
		opt(config)
	}
	rmqurl := fmt.Sprintf("%v:%d/api", config.Host, config.Port)
	restclient, err := rest.NewRestClient(rmqurl,
		rest.WithContext(config.Ctx),
		rest.WithAuthProvider(httpauthproviders.NewBasicAuthProvider(config.Username, config.Password)),
	)
	if err != nil {
		return nil, err
	}

	client := &RMQClient{
		restClient: restclient,
	}

	return client, nil
}

func (r *RMQClient) GetVhosts() ([]models.Vhost, error) {
	var vhosts []models.Vhost
	_, err := r.restClient.Get("/vhosts", &vhosts)
	if err != nil {
		return nil, err
	}
	return vhosts, nil
}

var (
	ErrInternalServerError = fmt.Errorf("internal server error")
	ErrVhostNotFound       = fmt.Errorf("vhost not found")
	ErrQueueNotFound       = fmt.Errorf("queue not found")
	ErrConnectionNotFound  = fmt.Errorf("connection not found")
	ErrChannelNotFound     = fmt.Errorf("channel not found")
	ErrNodeNotFound        = fmt.Errorf("node not found")
	ErrLimitsNotFound      = fmt.Errorf("limits not found")
)

func (r *RMQClient) Ping() error {
	_, err := r.restClient.Get("/overview", nil)
	if err != nil {
		return fmt.Errorf("%w. %w", ErrInternalServerError, err)
	}
	return nil
}

func (r *RMQClient) GetVhost(name string) (*models.Vhost, error) {

	var vhost models.Vhost
	status, err := r.restClient.Get("/vhosts/"+url.PathEscape(name), &vhost)
	if err != nil {
		switch status {
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrVhostNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}

	return &vhost, nil
}

func (r *RMQClient) GetConnections() ([]models.ConnectionResponse, error) {
	var connections []models.ConnectionResponse
	_, err := r.restClient.Get("/connections", &connections)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrConnectionNotFound, err)
	}
	return connections, nil
}

func (r *RMQClient) GetChannels() ([]models.ChannelResponse, error) {
	var channels []models.ChannelResponse
	_, err := r.restClient.Get("/channels", &channels)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrChannelNotFound, err)
	}
	return channels, nil
}

func (r *RMQClient) GetQueues() ([]models.RMQQueue, error) {
	var queues []models.RMQQueue
	_, err := r.restClient.Get("/queues", &queues)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrQueueNotFound, err)
	}
	return queues, nil
}

func (r *RMQClient) GetQueue(vhost string) ([]models.RMQQueue, error) {
	var queues []models.RMQQueue
	status, err := r.restClient.Get("/queues/"+url.PathEscape(vhost), &queues)
	if err != nil {
		switch status {
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrQueueNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return queues, nil
}

func (r *RMQClient) GetQueueByName(vhost string, name string) (*models.RMQQueue, error) {
	var queues models.RMQQueue
	status, err := r.restClient.Get("/queues/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), &queues)
	if err != nil {
		switch status {
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrQueueNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return &queues, nil
}

func (r *RMQClient) GetNodes() ([]models.RMQNode, error) {
	var nodes []models.RMQNode
	status, err := r.restClient.Get("/nodes", &nodes)
	if err != nil {
		switch status {
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrNodeNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return nodes, nil
}

func (r *RMQClient) GetMetrics(vhost string) (*models.VhostMetrics, error) {

	vhostObject, err := r.GetVhost(vhost)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve vhost %s. %w", vhost, err)
	}

	connections, err := r.GetConnections()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve connections for vhost %s. %w", vhost, err)
	}
	connCount := 0
	for _, c := range connections {
		if c.Vhost == vhost {
			connCount++
		}
	}

	channels, err := r.GetChannels()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channels for vhost %s. %w", vhost, err)
	}
	chanCount := 0
	for _, c := range channels {
		if c.Vhost == vhost {
			chanCount++
		}
	}

	queues, err := r.GetQueue(vhost)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve queues for vhost %s. %w", vhost, err)
	}

	return &models.VhostMetrics{
		Name:            vhost,
		Connections:     connCount,
		Channels:        chanCount,
		Queues:          len(queues),
		UnackedMessages: vhostObject.MessagesUnacknowledged,
		ReadyMessages:   vhostObject.Messages,
	}, nil
}

// TODO: Replace use of this with GetQueue(vhost)
// They serve the same purpose.
func (r *RMQClient) GetQueueDetails(vhost string) ([]models.QueueDetail, error) {

	queues, err := r.GetQueue(vhost)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve queues for vhost %s. %w", vhost, err)
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

func (r *RMQClient) GetVhostUsage(vhost string) (*models.VhostUsage, error) {

	queues, err := r.GetQueue(vhost)
	if err != nil {
		return nil, err
	}

	usage := &models.VhostUsage{
		Name:         vhost,
		MessageBytes: 0,
		DiskBytes:    0,
	}

	for _, q := range queues {
		usage.MessageBytes += q.MessageBytes
		usage.DiskBytes += q.MessageBytesPersistent
	}

	return usage, nil

}
func (r *RMQClient) GetLimits() ([]*models.RMQLimits, error) {
	var limits []*models.RMQLimits
	_, err := r.restClient.Get("/vhost-limits", &limits)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrLimitsNotFound, err)
	}
	return limits, nil
}

func (r *RMQClient) GetLimit(vhost string) (*models.RMQLimits, error) {
	var limit *models.RMQLimits
	_, err := r.restClient.Get("/vhost-limits/"+url.PathEscape(vhost), &limit)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrLimitsNotFound, err)
	}
	return limit, nil
}
