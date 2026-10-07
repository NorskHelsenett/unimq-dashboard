package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type RMQClientInterface interface {
	GetVhosts(filters ...Filter) ([]*models.Vhost, error)
	GetVhostUsage(filters ...Filter) ([]*models.RMQVhostUsage, error)
	GetVhostLimits(filters ...Filter) ([]*models.RMQVhostLimits, error)

	GetConnections(filters ...Filter) ([]*models.RMQConnection, error)
	GetChannels(filters ...Filter) ([]*models.RMQChannel, error)
	GetQueues(filters ...Filter) ([]*models.RMQQueue, error)

	GetNodes(filters ...Filter) ([]*models.RMQNode, error)

	GetMetrics(filters ...Filter) ([]*models.VhostMetrics, error)

	Ping() error
}

type RMQClient struct {
	restClient *rest.RestClient
}

// Guard against RMQClient not implementing RMQClientInterface in compile time.
var _ RMQClientInterface = (*RMQClient)(nil)

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

var (
	ErrInternalServerError = fmt.Errorf("internal server error")
	ErrVhostNotFound       = fmt.Errorf("vhost not found")
	ErrQueueNotFound       = fmt.Errorf("queue not found")
	ErrConnectionNotFound  = fmt.Errorf("connection not found")
	ErrChannelNotFound     = fmt.Errorf("channel not found")
	ErrNodeNotFound        = fmt.Errorf("node not found")
	ErrLimitsNotFound      = fmt.Errorf("limits not found")
)

func (r *RMQClient) GetVhosts(filters ...Filter) ([]*models.Vhost, error) {

	uri, vhostsFilter, err := convertFiltersToPathParams(filters, "/vhosts")
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
	}

	var vhosts []*models.Vhost
	status, err := r.restClient.Get(uri, &vhosts)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w. %w", ErrVhostNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}

	if len(vhostsFilter) > 0 {
		filteredVhosts := make([]*models.Vhost, 0)
		for _, v := range vhosts {
			if slices.Contains(vhostsFilter, v.Name) {
				filteredVhosts = append(filteredVhosts, v)
			}
		}
		vhosts = filteredVhosts
	}

	return vhosts, nil
}

func (r *RMQClient) GetConnections(filters ...Filter) ([]*models.RMQConnection, error) {
	queryparams := convertFiltersToQueryParams(filters)

	var connections []*models.RMQConnection
	uri := "/connections" + queryparams
	status, err := r.restClient.Get(uri, &connections)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w. %w", ErrConnectionNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return connections, nil
}

func (r *RMQClient) GetChannels(filters ...Filter) ([]*models.RMQChannel, error) {
	queryparams := convertFiltersToQueryParams(filters)
	var channels []*models.RMQChannel
	uri := "/channels" + queryparams
	status, err := r.restClient.Get(uri, &channels)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w. %w", ErrChannelNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return channels, nil
}

func (r *RMQClient) GetQueues(filters ...Filter) ([]*models.RMQQueue, error) {
	queryparams := convertFiltersToQueryParams(filters)
	var queues []*models.RMQQueue
	uri := "/queues" + queryparams
	status, err := r.restClient.Get(uri, &queues)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w. %w", ErrQueueNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}

	return queues, nil

}
func (r *RMQClient) GetNodes(filters ...Filter) ([]*models.RMQNode, error) {
	queryparams := convertFiltersToQueryParams(filters)
	var nodes []*models.RMQNode
	uri := "/nodes" + queryparams
	status, err := r.restClient.Get(uri, &nodes)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w. %w", ErrNodeNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return nodes, nil
}

func (r *RMQClient) GetVhostLimits(filters ...Filter) ([]*models.RMQVhostLimits, error) {
	uri, vhosts, err := convertFiltersToPathParams(filters, "/vhost-limits")
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
	}
	var limits []*models.RMQVhostLimits
	status, err := r.restClient.Get(uri, &limits)
	if err != nil {
		switch status {
		case http.StatusNotFound:
			return nil, fmt.Errorf("%w. %w", ErrLimitsNotFound, err)

		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}

	if len(vhosts) > 0 {
		filteredLimits := make([]*models.RMQVhostLimits, 0)
		for _, l := range limits {
			if slices.Contains(vhosts, l.Vhost) {
				filteredLimits = append(filteredLimits, l)
			}
		}
		limits = filteredLimits
	}
	return limits, nil
}

func (r *RMQClient) GetMetrics(filters ...Filter) ([]*models.VhostMetrics, error) {

	vhostFilter := getFiltersByType(filters, FilterTypeVhost)
	vhostNames := make([]string, 0, len(vhostFilter))
	for _, v := range vhostFilter {
		vhostNames = append(vhostNames, v.Value)
	}
	vhostObject, err := r.GetVhosts(vhostFilter...)
	if err != nil {
		if errors.Is(err, ErrVhostNotFound) {
			return nil, err
		}
		if len(vhostFilter) > 0 {
			return nil, fmt.Errorf("failed to retrieve vhost(s) %v. %w", vhostNames, err)
		}
		return nil, fmt.Errorf("failed to retrieve any vhosts. %w", err)
	}

	connectionFilter := getFiltersByType(filters, FilterTypeConnection)
	connectionFilter = append(connectionFilter, vhostFilter...)
	connections, err := r.GetConnections(connectionFilter...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve connections for vhost(s) %v. %w", vhostNames, err)
	}

	channelFilter := getFiltersByType(filters, FilterTypeChannel)
	channelFilter = append(channelFilter, vhostFilter...)
	channels, err := r.GetChannels(channelFilter...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve channels for vhost %v. %w", vhostNames, err)
	}

	queueFilter := getFiltersByType(filters, FilterTypeQueue)
	queueFilter = append(queueFilter, vhostFilter...)
	queues, err := r.GetQueues(queueFilter...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve queues for vhost %v. %w", vhostNames, err)
	}

	metrics := make([]*models.VhostMetrics, 0, len(vhostObject))

	for _, v := range vhostObject {
		vhostMetrics := &models.VhostMetrics{
			Name:            v.Name,
			Connections:     0,
			Channels:        0,
			Queues:          0,
			UnackedMessages: 0,
		}

		for _, q := range queues {
			if q.Vhost == v.Name {
				vhostMetrics.UnackedMessages += q.MessagesUnacknowledged
				vhostMetrics.Queues++
			}
		}

		for _, c := range connections {
			if c.Vhost == v.Name {
				vhostMetrics.Connections++
			}
		}

		for _, c := range channels {
			if c.Vhost == v.Name {
				vhostMetrics.Channels++
			}
		}

		metrics = append(metrics, vhostMetrics)
	}

	return metrics, nil
}

func (r *RMQClient) Ping() error {
	_, err := r.restClient.Get("/overview", nil)
	if err != nil {
		return fmt.Errorf("%w. %w", ErrInternalServerError, err)
	}
	return nil
}

func (r *RMQClient) GetVhostUsage(filters ...Filter) ([]*models.RMQVhostUsage, error) {

	vhostFilter := getFiltersByType(filters, FilterTypeVhost)
	queueFilter := getFiltersByType(filters, FilterTypeQueue)
	filter := make([]Filter, 0, len(vhostFilter)+len(queueFilter))
	filter = append(filter, vhostFilter...)
	filter = append(filter, queueFilter...)
	queues, err := r.GetQueues(filter...)
	if err != nil {
		return nil, err
	}

	usage := make([]*models.RMQVhostUsage, 0, len(vhostFilter))
	for _, v := range vhostFilter {
		use := &models.RMQVhostUsage{
			Name:         v.Value,
			MessageBytes: 0,
			DiskBytes:    0,
		}

		for _, q := range queues {
			use.MessageBytes += q.MessageBytes
			use.DiskBytes += q.MessageBytesPersistent
		}
		usage = append(usage, use)
	}

	return usage, nil

}
