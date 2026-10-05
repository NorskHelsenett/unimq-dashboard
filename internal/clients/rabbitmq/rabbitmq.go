package rabbitmq

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type RMQClientInterface interface {
	GetVhosts(filters ...Filter) ([]models.Vhost, error)

	GetConnections(filters ...Filter) ([]*models.RMQConnection, error)
	GetChannels(filters ...Filter) ([]*models.RMQChannel, error)
	GetQueues(filters ...Filter) ([]*models.RMQQueue, error)

	GetNodes(filters ...Filter) ([]*models.RMQNode, error)

	GetMetrics(filters ...Filter) ([]*models.VhostMetrics, error)
	GetVhostUsage(filters ...Filter) ([]*models.RMQVhostUsage, error)
}

type (
	Filter struct {
		Parameter Parameter
		Value     string
	}

	Parameter      string
	FilterProperty string
)

var (
	ErrUnsupportedFilterParameter = fmt.Errorf("unsupported filter parameter")
)

func convertFiltersToQueryParams(filters []Filter) string {
	builder := &strings.Builder{}

	builder.WriteString("?")

	for i, filter := range filters {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(string(filter.Parameter))
		builder.WriteString("=")
		builder.WriteString(url.QueryEscape(filter.Value))
	}

	return builder.String()
}

const (
	ParameterName       Parameter = "name"
	ParameterPage       Parameter = "page"
	ParameterPageSize   Parameter = "page_size"
	ParameterUseRegex   Parameter = "use_regex"
	ParameterPagination Parameter = "pagination"
)

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

func (r *RMQClient) GetVhosts(filters ...Filter) ([]*models.Vhost, error) {
	var vhost string
	vhostsFilter := make([]string, 0)

	switch {
	case len(filters) == 0:
		//  No Filters, fetch all.
	case len(filters) == 1:
		// One filter, check if it's the name filter, and use the path parameter to fetch the specific vhost.
		if filters[0].Parameter == ParameterName {
			vhost = filters[0].Value
		} else {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedFilterParameter, filters[0].Parameter)
		}
	case len(filters) > 1:
		// Multiple filters, check if the name filter is present, query for all vhosts and filter the results based on the name filter.
		for _, filter := range filters {
			if filter.Parameter == ParameterName {
				vhostsFilter = append(vhostsFilter, filter.Value)
			} else {
				return nil, fmt.Errorf("%w: %s", ErrUnsupportedFilterParameter, filter.Parameter)
			}
		}
	default:
		return nil, fmt.Errorf("only one filter is supported for GetVhosts")
	}

	uri := "/vhosts"
	if vhost != "" {
		uri += "/" + vhost
	}

	var vhosts []*models.Vhost
	_, err := r.restClient.Get(uri, &vhosts)
	if err != nil {
		return nil, err
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

func (r *RMQClient) GetConnections() ([]models.RMQConnection, error) {
	var connections []models.RMQConnection
	_, err := r.restClient.Get("/connections", &connections)
	if err != nil {
		return nil, fmt.Errorf("%w. %w", ErrConnectionNotFound, err)
	}
	return connections, nil
}

func (r *RMQClient) GetChannels() ([]models.RMQChannel, error) {
	var channels []models.RMQChannel
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

func (r *RMQClient) GetVhostUsage(vhost string) (*models.RMQVhostUsage, error) {

	queues, err := r.GetQueue(vhost)
	if err != nil {
		return nil, err
	}

	usage := &models.RMQVhostUsage{
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
func (r *RMQClient) GetVhostLimits() ([]*models.RMQVhostLimits, error) {
	var limits []*models.RMQVhostLimits
	status, err := r.restClient.Get("/vhost-limits", &limits)
	if err != nil {
		switch status {
		case 404:
			return nil, fmt.Errorf("%w. %w", ErrLimitsNotFound, err)

		default:
			return nil, fmt.Errorf("%w. %w", ErrInternalServerError, err)
		}
	}
	return limits, nil
}

func (r *RMQClient) GetVhostLimit(vhost string) (*models.RMQVhostLimits, error) {
	var limit *models.RMQVhostLimits
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
