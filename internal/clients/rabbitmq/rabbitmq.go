package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type RMQClientInterface interface {
	GetVhosts(filters ...Filter) ([]models.Vhost, error)
	GetVhostUsage(filters ...Filter) ([]*models.RMQVhostUsage, error)
	GetVhostLimits(filters ...Filter) ([]*models.RMQVhostLimits, error)

	GetConnections(filters ...Filter) ([]*models.RMQConnection, error)
	GetChannels(filters ...Filter) ([]*models.RMQChannel, error)
	GetQueues(filters ...Filter) ([]*models.RMQQueue, error)

	GetNodes(filters ...Filter) ([]*models.RMQNode, error)

	GetMetrics(filters ...Filter) ([]*models.VhostMetrics, error)
}

type RMQClient struct {
	restClient *rest.RestClient
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

	var vhosts []*models.Vhost
	status, err := r.restClient.Get(uri, &vhosts)
	if err != nil {
		switch status {
		case 404:
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
	_, err := r.restClient.Get(uri, &connections)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrConnectionNotFound, err)
	}
	return connections, nil
}

func (r *RMQClient) GetChannels(filters ...Filter) ([]*models.RMQChannel, error) {
	queryparams := convertFiltersToQueryParams(filters)
	var channels []*models.RMQChannel
	uri := "/channels" + queryparams
	_, err := r.restClient.Get(uri, &channels)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrChannelNotFound, err)
	}
	return channels, nil
}

func (r *RMQClient) GetQueues(filters ...Filter) ([]*models.RMQQueue, error) {
	queryparams := convertFiltersToQueryParams(filters)
	var queues []*models.RMQQueue
	uri := "/queues" + queryparams
	_, err := r.restClient.Get(uri, &queues)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrQueueNotFound, err)
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
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrNodeNotFound, err)
		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return nodes, nil
}

func (r *RMQClient) GetMetrics(filters ...Filter) ([]*models.VhostMetrics, error) {

	vhostFilter := getFiltersByType(filters, FilterTypeVhost)
	vhostObject, err := r.GetVhosts(vhostFilter...)
	if err != nil {
		if errors.Is(err, ErrVhostNotFound) {
			return nil, fmt.Errorf("%w. %w", ErrVhostNotFound, err)
		}
		if len(vhostFilter) > 0 {
			vhostNames := make([]string, 0, len(vhostObject))
			for _, v := range vhostObject {
				vhostNames = append(vhostNames, v.Name)
			}
			return nil, fmt.Errorf("failed to retrieve vhost(s) %v. %w", vhostNames, err)
		}
		return nil, fmt.Errorf("failed to retrieve any vhosts. %w", err)
	}

	connectionFilter := getFiltersByType(filters, FilterTypeConnection)
	connectionFilter = append(connectionFilter, vhostFilter...)
	connections, err := r.GetConnections(connectionFilter...)
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

	queues, err := r.GetQueues(vhostFilter...)
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

func (r *RMQClient) Ping() error {
	_, err := r.restClient.Get("/overview", nil)
	if err != nil {
		return fmt.Errorf("%w. %w", ErrInternalServerError, err)
	}
	return nil
}

// TODO: Convert to
// func (r *RMQClient) GetVhostUsage(vhost string) (*models.RMQVhostUsage, error) {
//
// 	queues, err := r.GetQueue(vhost)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	usage := &models.RMQVhostUsage{
// 		Name:         vhost,
// 		MessageBytes: 0,
// 		DiskBytes:    0,
// 	}
//
// 	for _, q := range queues {
// 		usage.MessageBytes += q.MessageBytes
// 		usage.DiskBytes += q.MessageBytesPersistent
// 	}
//
// 	return usage, nil
//
// }

func (r *RMQClient) GetVhostLimits(filters ...Filter) ([]*models.RMQVhostLimits, error) {
	var limits []*models.RMQVhostLimits
	status, err := r.restClient.Get("/vhost-limits/"+url.PathEscape(vhost), &limit)
	if err != nil {
		switch status {
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrLimitsNotFound, err)

		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return limit, nil
}
