package rmq_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq/rmqfake"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	adminGroup = "unimq-admins"
	testVhost  = "team-a"
)

// requestWithGroups builds a request carrying the claims the handlers expect,
// plus any chi URL parameters the route would normally supply.
func requestWithGroups(t *testing.T, groups []string, urlParams map[string]string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	anyGroups := make([]any, 0, len(groups))
	for _, g := range groups {
		anyGroups = append(anyGroups, g)
	}
	claims := map[string]any{"groups": anyGroups}

	ctx := context.WithValue(req.Context(), httpsuite.ClaimsContextKey, claims)

	if len(urlParams) > 0 {
		routeCtx := chi.NewRouteContext()
		for k, v := range urlParams {
			routeCtx.URLParams.Add(k, v)
		}
		ctx = context.WithValue(ctx, chi.RouteCtxKey, routeCtx)
	}

	return req.WithContext(ctx)
}

// decodeBody unwraps the {code,message,body} envelope SendResponse writes.
func decodeBody[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var envelope struct {
		Body T `json:"body"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &envelope), "body: %s", rec.Body.String())
	return envelope.Body
}

func TestNewRMQHandlerAcceptsAnInterfaceImplementation(t *testing.T) {
	db := &database.Database{}
	handler := rmq.NewRMQHandler(&rmqfake.Client{}, db, []string{adminGroup})
	require.NotNil(t, handler)
	assert.NotNil(t, handler.RMQClient)
}

func TestGetVhostsHandlerReturnsVhosts(t *testing.T) {
	client := &rmqfake.Client{
		GetVhostsFunc: func(_ ...rabbitmq.Filter) ([]*models.Vhost, error) {
			return []*models.Vhost{{Name: "/"}, {Name: testVhost}}, nil
		},
	}
	db := &database.Database{}
	handler := rmq.NewRMQHandler(client, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetVhostsHandler(rec, requestWithGroups(t, []string{adminGroup}, nil))

	require.Equal(t, http.StatusOK, rec.Code)
	vhosts := decodeBody[[]*models.Vhost](t, rec)
	require.Len(t, vhosts, 2)
	assert.Equal(t, "/", vhosts[0].Name)
	assert.Len(t, client.CallsTo("GetVhosts"), 1)
}

func TestGetVhostsHandlerForbiddenWithoutAdminGroup(t *testing.T) {
	client := &rmqfake.Client{}
	db := &database.Database{}
	handler := rmq.NewRMQHandler(client, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetVhostsHandler(rec, requestWithGroups(t, []string{"some-other-group"}, nil))

	require.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, client.Calls(), "the client must not be hit when authorisation fails")
}

func TestGetVhostsHandlerPropagatesClientError(t *testing.T) {
	db := &database.Database{}
	handler := rmq.NewRMQHandler(&rmqfake.Client{
		GetVhostsFunc: func(_ ...rabbitmq.Filter) ([]*models.Vhost, error) {
			return nil, fmt.Errorf("%w. boom", rabbitmq.ErrInternalServerError)
		},
	}, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetVhostsHandler(rec, requestWithGroups(t, []string{adminGroup}, nil))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetVhostHandlerPassesNameFilter(t *testing.T) {
	db := &database.Database{}
	client := &rmqfake.Client{
		GetVhostsFunc: func(_ ...rabbitmq.Filter) ([]*models.Vhost, error) {
			return []*models.Vhost{{Name: testVhost}}, nil
		},
	}
	handler := rmq.NewRMQHandler(client, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetVhostHandler(rec, requestWithGroups(t, []string{adminGroup},
		map[string]string{"vhost-name": testVhost}))

	require.Equal(t, http.StatusOK, rec.Code)

	calls := client.CallsTo("GetVhosts")
	require.Len(t, calls, 1)
	require.Len(t, calls[0].Filters, 1)
	assert.Equal(t, rabbitmq.NewFilter(rabbitmq.ParameterName, rabbitmq.FilterTypeVhost, testVhost), calls[0].Filters[0])
}

func TestGetVhostLimitsHandlerReturnsEmptyLimitsWhenNoneDefined(t *testing.T) {
	db := &database.Database{}
	handler := rmq.NewRMQHandler(&rmqfake.Client{
		GetVhostLimitsFunc: func(_ ...rabbitmq.Filter) ([]*models.RMQVhostLimits, error) {
			return nil, fmt.Errorf("%w. not set", rabbitmq.ErrLimitsNotFound)
		},
	}, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetVhostLimitsHandler(rec, requestWithGroups(t, []string{adminGroup},
		map[string]string{"vhost-name": testVhost}))

	require.Equal(t, http.StatusOK, rec.Code, "a vhost without limits is not an error")
	limits := decodeBody[*models.RMQVhostLimits](t, rec)
	require.NotNil(t, limits)
	assert.Equal(t, testVhost, limits.Vhost)
	assert.Equal(t, 0, limits.MaxConnections)
}

func TestGetVhostLimitsHandlerSurfacesOtherErrors(t *testing.T) {
	db := &database.Database{}
	handler := rmq.NewRMQHandler(&rmqfake.Client{
		GetVhostLimitsFunc: func(_ ...rabbitmq.Filter) ([]*models.RMQVhostLimits, error) {
			return nil, errors.New("connection refused")
		},
	}, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetVhostLimitsHandler(rec, requestWithGroups(t, []string{adminGroup},
		map[string]string{"vhost-name": testVhost}))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetQueuesByNameHandlerMapsNotFound(t *testing.T) {
	db := &database.Database{}
	handler := rmq.NewRMQHandler(&rmqfake.Client{
		GetQueuesFunc: func(_ ...rabbitmq.Filter) ([]*models.RMQQueue, error) {
			return nil, fmt.Errorf("%w. missing", rabbitmq.ErrQueueNotFound)
		},
	}, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetQueuesByNameHandler(rec, requestWithGroups(t, []string{adminGroup},
		map[string]string{"vhost-name": testVhost, "queue-id": "q1"}))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetRMQNodesHandler(t *testing.T) {
	db := &database.Database{}
	handler := rmq.NewRMQHandler(&rmqfake.Client{
		GetNodesFunc: func(_ ...rabbitmq.Filter) ([]*models.RMQNode, error) {
			return []*models.RMQNode{{Name: "rabbit@node1", MemUsed: 100}}, nil
		},
	}, db, []string{adminGroup})

	rec := httptest.NewRecorder()
	handler.GetRMQNodesHandler(rec, requestWithGroups(t, []string{adminGroup}, nil))

	require.Equal(t, http.StatusOK, rec.Code)
	nodes := decodeBody[[]*models.RMQNode](t, rec)
	require.Len(t, nodes, 1)
	assert.Equal(t, "rabbit@node1", nodes[0].Name)
}
