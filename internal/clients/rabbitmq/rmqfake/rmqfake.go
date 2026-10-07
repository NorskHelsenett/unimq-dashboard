// Package rmqfake provides an in-memory implementation of
// [rabbitmq.RMQClientInterface] for use in tests.
//
// Every method is backed by an optional function field. When a field is nil the
// method returns an empty result and a nil error, so a test only has to set the
// behaviour it actually cares about:
//
//	client := &rmqfake.Client{
//		GetVhostsFunc: func(_ ...rabbitmq.Filter) ([]*models.Vhost, error) {
//			return []*models.Vhost{{Name: "/"}}, nil
//		},
//	}
//	handler := rmq.NewRMQHandler(client, []string{"admins"})
//
// Calls are recorded so tests can assert on the filters that were passed.
package rmqfake

import (
	"sync"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

// Call records a single invocation of a Client method.
type Call struct {
	Method  string
	Filters []rabbitmq.Filter
}

// Client is a configurable test double for rabbitmq.RMQClientInterface.
// The zero value is usable and returns empty results for every method.
type Client struct {
	GetVhostsFunc      func(filters ...rabbitmq.Filter) ([]*models.Vhost, error)
	GetVhostUsageFunc  func(filters ...rabbitmq.Filter) ([]*models.RMQVhostUsage, error)
	GetVhostLimitsFunc func(filters ...rabbitmq.Filter) ([]*models.RMQVhostLimits, error)
	GetConnectionsFunc func(filters ...rabbitmq.Filter) ([]*models.RMQConnection, error)
	GetChannelsFunc    func(filters ...rabbitmq.Filter) ([]*models.RMQChannel, error)
	GetQueuesFunc      func(filters ...rabbitmq.Filter) ([]*models.RMQQueue, error)
	GetNodesFunc       func(filters ...rabbitmq.Filter) ([]*models.RMQNode, error)
	GetMetricsFunc     func(filters ...rabbitmq.Filter) ([]*models.VhostMetrics, error)
	PingFunc           func() error

	mu    sync.Mutex
	calls []Call
}

// Guard against Client not implementing RMQClientInterface at compile time.
var _ rabbitmq.RMQClientInterface = (*Client)(nil)

func (c *Client) record(method string, filters []rabbitmq.Filter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, Call{Method: method, Filters: filters})
}

// Calls returns a copy of every recorded invocation, in call order.
func (c *Client) Calls() []Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Call(nil), c.calls...)
}

// CallsTo returns the recorded invocations of a single method.
func (c *Client) CallsTo(method string) []Call {
	calls := make([]Call, 0)
	for _, call := range c.Calls() {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// Reset clears the recorded calls.
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = nil
}

func (c *Client) GetVhosts(filters ...rabbitmq.Filter) ([]*models.Vhost, error) {
	c.record("GetVhosts", filters)
	if c.GetVhostsFunc == nil {
		return []*models.Vhost{}, nil
	}
	return c.GetVhostsFunc(filters...)
}

func (c *Client) GetVhostUsage(filters ...rabbitmq.Filter) ([]*models.RMQVhostUsage, error) {
	c.record("GetVhostUsage", filters)
	if c.GetVhostUsageFunc == nil {
		return []*models.RMQVhostUsage{}, nil
	}
	return c.GetVhostUsageFunc(filters...)
}

func (c *Client) GetVhostLimits(filters ...rabbitmq.Filter) ([]*models.RMQVhostLimits, error) {
	c.record("GetVhostLimits", filters)
	if c.GetVhostLimitsFunc == nil {
		return []*models.RMQVhostLimits{}, nil
	}
	return c.GetVhostLimitsFunc(filters...)
}

func (c *Client) GetConnections(filters ...rabbitmq.Filter) ([]*models.RMQConnection, error) {
	c.record("GetConnections", filters)
	if c.GetConnectionsFunc == nil {
		return []*models.RMQConnection{}, nil
	}
	return c.GetConnectionsFunc(filters...)
}

func (c *Client) GetChannels(filters ...rabbitmq.Filter) ([]*models.RMQChannel, error) {
	c.record("GetChannels", filters)
	if c.GetChannelsFunc == nil {
		return []*models.RMQChannel{}, nil
	}
	return c.GetChannelsFunc(filters...)
}

func (c *Client) GetQueues(filters ...rabbitmq.Filter) ([]*models.RMQQueue, error) {
	c.record("GetQueues", filters)
	if c.GetQueuesFunc == nil {
		return []*models.RMQQueue{}, nil
	}
	return c.GetQueuesFunc(filters...)
}

func (c *Client) GetNodes(filters ...rabbitmq.Filter) ([]*models.RMQNode, error) {
	c.record("GetNodes", filters)
	if c.GetNodesFunc == nil {
		return []*models.RMQNode{}, nil
	}
	return c.GetNodesFunc(filters...)
}

func (c *Client) GetMetrics(filters ...rabbitmq.Filter) ([]*models.VhostMetrics, error) {
	c.record("GetMetrics", filters)
	if c.GetMetricsFunc == nil {
		return []*models.VhostMetrics{}, nil
	}
	return c.GetMetricsFunc(filters...)
}

func (c *Client) Ping() error {
	c.record("Ping", nil)
	if c.PingFunc == nil {
		return nil
	}
	return c.PingFunc()
}
