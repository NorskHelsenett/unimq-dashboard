package rabbitmq

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordedRequest captures what the client actually put on the wire, including the
// raw (un-decoded) path so escaping regressions are caught.
type recordedRequest struct {
	RawPath  string
	RawQuery string
}

// newTestClient starts an httptest server and returns a client pointed at it via
// NewRMQClient, so the host/port/base-URL assembly is exercised too.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*RMQClient, *recordedRequest) {
	t.Helper()

	recorded := &recordedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded.RawPath = r.URL.EscapedPath()
		recorded.RawQuery = r.URL.RawQuery
		handler(w, r)
	}))
	t.Cleanup(srv.Close)

	host, portStr, found := strings.Cut(strings.TrimPrefix(srv.URL, "http://"), ":")
	require.True(t, found, "unexpected httptest URL %q", srv.URL)
	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	client, err := NewRMQClient(
		WithRMQHost("http://"+host),
		WithRMQPort(port),
		WithRMQUsername("guest"),
		WithRMQPassword("guest"),
	)
	require.NoError(t, err)

	return client, recorded
}

func jsonHandler(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestGetVhostsUnfiltered(t *testing.T) {
	client, recorded := newTestClient(t, jsonHandler(http.StatusOK,
		`[{"name":"/"},{"name":"other"}]`))

	vhosts, err := client.GetVhosts()

	require.NoError(t, err)
	require.Len(t, vhosts, 2)
	assert.Equal(t, "/", vhosts[0].Name)
	assert.Equal(t, "/api/vhosts", recorded.RawPath)
	assert.Empty(t, recorded.RawQuery, "no filters must not produce a bare '?'")
}

func TestGetVhostsFiltersClientSideForMultipleNames(t *testing.T) {
	client, recorded := newTestClient(t, jsonHandler(http.StatusOK,
		`[{"name":"/"},{"name":"keep"},{"name":"drop"}]`))

	vhosts, err := client.GetVhosts(
		NewFilter(ParameterName, FilterTypeVhost, "/"),
		NewFilter(ParameterName, FilterTypeVhost, "keep"),
	)

	require.NoError(t, err)
	require.Len(t, vhosts, 2)
	assert.Equal(t, "/", vhosts[0].Name)
	assert.Equal(t, "keep", vhosts[1].Name)
	assert.Equal(t, "/api/vhosts", recorded.RawPath, "multiple names are fetched in bulk then filtered")
}

func TestGetVhostsNotFound(t *testing.T) {
	client, _ := newTestClient(t, jsonHandler(http.StatusNotFound, `{"error":"Object Not Found"}`))

	vhosts, err := client.GetVhosts()

	require.ErrorIs(t, err, ErrVhostNotFound)
	assert.Nil(t, vhosts)
}

func TestGetVhostsServerError(t *testing.T) {
	client, _ := newTestClient(t, jsonHandler(http.StatusInternalServerError, `{"error":"boom"}`))

	_, err := client.GetVhosts()

	require.ErrorIs(t, err, ErrInternalServerError)
	assert.NotErrorIs(t, err, ErrVhostNotFound)
}

func TestGetQueues(t *testing.T) {
	client, recorded := newTestClient(t, jsonHandler(http.StatusOK,
		`[{"name":"q1","vhost":"/","messages_unacknowledged":3}]`))

	queues, err := client.GetQueues(NewFilter(ParameterName, FilterTypeQueue, "q1"))

	require.NoError(t, err)
	require.Len(t, queues, 1)
	assert.Equal(t, "q1", queues[0].Name)
	assert.Equal(t, "/api/queues", recorded.RawPath)
	assert.Equal(t, "name=q1", recorded.RawQuery)
}

func TestGetQueuesNotFound(t *testing.T) {
	client, _ := newTestClient(t, jsonHandler(http.StatusNotFound, `{"error":"Object Not Found"}`))

	_, err := client.GetQueues()

	require.ErrorIs(t, err, ErrQueueNotFound)
}

func TestGetConnectionsAndChannelsErrorMapping(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		call     func(c *RMQClient) error
		expected error
	}{
		{
			name:     "connections not found",
			status:   http.StatusNotFound,
			call:     func(c *RMQClient) error { _, err := c.GetConnections(); return err },
			expected: ErrConnectionNotFound,
		},
		{
			name:     "connections server error",
			status:   http.StatusBadGateway,
			call:     func(c *RMQClient) error { _, err := c.GetConnections(); return err },
			expected: ErrInternalServerError,
		},
		{
			name:     "channels not found",
			status:   http.StatusNotFound,
			call:     func(c *RMQClient) error { _, err := c.GetChannels(); return err },
			expected: ErrChannelNotFound,
		},
		{
			name:     "nodes not found",
			status:   http.StatusNotFound,
			call:     func(c *RMQClient) error { _, err := c.GetNodes(); return err },
			expected: ErrNodeNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			client, _ := newTestClient(t, jsonHandler(tc.status, `{"error":"nope"}`))
			require.ErrorIs(t, tc.call(client), tc.expected)
		})
	}
}

func TestGetVhostLimitsUsesEscapedPathParameter(t *testing.T) {
	client, recorded := newTestClient(t, jsonHandler(http.StatusOK,
		`[{"vhost":"/","values":{"max_connections":10,"max_queues":5}}]`))

	limits, err := client.GetVhostLimits(NewFilter(ParameterName, FilterTypeVhost, "/"))

	require.NoError(t, err)
	require.Len(t, limits, 1)
	assert.Equal(t, 10, limits[0].MaxConnections)
	assert.Equal(t, 5, limits[0].MaxQueues)
	assert.Equal(t, "/api/vhost-limits/%2F", recorded.RawPath, "the default vhost must stay percent-encoded")
}

func TestGetVhostLimitsRejectsUnsupportedFilter(t *testing.T) {
	client, _ := newTestClient(t, jsonHandler(http.StatusOK, `[]`))

	_, err := client.GetVhostLimits(NewFilter(ParameterPageSize, FilterTypeVhost, "10"))

	require.ErrorIs(t, err, ErrInternalServerError)
	assert.ErrorIs(t, err, ErrUnsupportedFilterParameter)
}

func TestGetVhostLimitsNotFound(t *testing.T) {
	client, _ := newTestClient(t, jsonHandler(http.StatusNotFound, `{"error":"Object Not Found"}`))

	_, err := client.GetVhostLimits()

	require.ErrorIs(t, err, ErrLimitsNotFound)
}

func TestGetMetricsAggregatesPerVhost(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body string
		switch r.URL.EscapedPath() {
		case "/api/vhosts":
			body = `[{"name":"/"},{"name":"other"}]`
		case "/api/connections":
			body = `[{"vhost":"/"},{"vhost":"/"},{"vhost":"other"}]`
		case "/api/channels":
			body = `[{"vhost":"/"}]`
		case "/api/queues":
			body = `[{"name":"q1","vhost":"/","messages_unacknowledged":2},
			          {"name":"q2","vhost":"/","messages_unacknowledged":3},
			          {"name":"q3","vhost":"other","messages_unacknowledged":7}]`
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	})

	metrics, err := client.GetMetrics()

	require.NoError(t, err)
	require.Len(t, metrics, 2)

	assert.Equal(t, "/", metrics[0].Name)
	assert.Equal(t, 2, metrics[0].Connections)
	assert.Equal(t, 1, metrics[0].Channels)
	assert.Equal(t, 2, metrics[0].Queues)
	assert.Equal(t, 5, metrics[0].UnackedMessages)

	assert.Equal(t, "other", metrics[1].Name)
	assert.Equal(t, 1, metrics[1].Connections)
	assert.Equal(t, 0, metrics[1].Channels)
	assert.Equal(t, 1, metrics[1].Queues)
	assert.Equal(t, 7, metrics[1].UnackedMessages)
}

func TestGetMetricsPropagatesVhostLookupFailure(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() == "/api/vhosts" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`[]`))
	})

	_, err := client.GetMetrics()

	require.ErrorIs(t, err, ErrVhostNotFound)
}

func TestPing(t *testing.T) {
	client, recorded := newTestClient(t, jsonHandler(http.StatusOK, `{"product_name":"RabbitMQ"}`))

	require.NoError(t, client.Ping())
	assert.Equal(t, "/api/overview", recorded.RawPath)
}

func TestPingFailure(t *testing.T) {
	client, _ := newTestClient(t, jsonHandler(http.StatusUnauthorized, `{"error":"not_authorised"}`))

	require.ErrorIs(t, client.Ping(), ErrInternalServerError)
}

func TestClientSendsBasicAuthAndUserAgent(t *testing.T) {
	var gotUser, gotPass, gotAgent string
	var gotOK bool

	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, gotOK = r.BasicAuth()
		gotAgent = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`[]`))
	})

	require.NoError(t, client.Ping())
	assert.True(t, gotOK, "expected basic auth to be set")
	assert.Equal(t, "guest", gotUser)
	assert.Equal(t, "guest", gotPass)
	assert.Equal(t, "UniMQ", gotAgent)
}
