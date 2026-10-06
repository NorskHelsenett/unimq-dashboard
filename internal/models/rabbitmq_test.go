package models_test

import (
	"encoding/json"
	"testing"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRMQVhostLimits_UnmarshalJSON_ValuesPresent(t *testing.T) {
	data := []byte(`{"vhost":"/","values":{"max_connections":100,"max_queues":50}}`)

	var limits models.RMQVhostLimits
	err := json.Unmarshal(data, &limits)
	require.NoError(t, err)

	assert.Equal(t, "/", limits.Vhost)
	assert.Equal(t, 100, limits.MaxConnections)
	assert.Equal(t, 50, limits.MaxQueues)
}

func TestRMQVhostLimits_UnmarshalJSON_ValuesNull(t *testing.T) {
	// RabbitMQ reports `null` for an unset limit rather than omitting it.
	data := []byte(`{"vhost":"/","values":{"max_connections":null,"max_queues":null}}`)

	var limits models.RMQVhostLimits
	err := json.Unmarshal(data, &limits)
	require.NoError(t, err)

	assert.Equal(t, 0, limits.MaxConnections)
	assert.Equal(t, 0, limits.MaxQueues)
}

func TestRMQVhostLimits_UnmarshalJSON_ValuesMissing(t *testing.T) {
	// No limits configured at all: RabbitMQ omits "values" entirely.
	data := []byte(`{"vhost":"/"}`)

	var limits models.RMQVhostLimits
	err := json.Unmarshal(data, &limits)
	require.NoError(t, err)

	assert.Equal(t, "/", limits.Vhost)
	assert.Equal(t, 0, limits.MaxConnections)
	assert.Equal(t, 0, limits.MaxQueues)
}

func TestRMQVhostLimits_UnmarshalJSON_MalformedJSON(t *testing.T) {
	var limits models.RMQVhostLimits
	err := json.Unmarshal([]byte(`{"vhost":`), &limits)
	assert.Error(t, err)
}

func TestRMQQueue_UnmarshalJSON_FullMessageStats(t *testing.T) {
	data := []byte(`{
		"name": "my-queue",
		"vhost": "/",
		"messages": 10,
		"messages_unacknowledged": 2,
		"consumers": 1,
		"message_bytes": 1024,
		"message_bytes_persistent": 512,
		"message_stats": {
			"publish_details": {"rate": 1.5},
			"deliver_get_details": {"rate": 2.5},
			"redeliver_details": {"rate": 0.1}
		}
	}`)

	var q models.RMQQueue
	err := json.Unmarshal(data, &q)
	require.NoError(t, err)

	assert.Equal(t, "my-queue", q.Name)
	assert.Equal(t, "/", q.Vhost)
	assert.Equal(t, 10, q.Messages)
	assert.Equal(t, 2, q.MessagesUnacknowledged)
	assert.Equal(t, 1, q.Consumers)
	assert.Equal(t, int64(1024), q.MessageBytes)
	assert.Equal(t, int64(512), q.MessageBytesPersistent)
	assert.InDelta(t, 1.5, q.PublishRate, 0.0001)
	assert.InDelta(t, 2.5, q.DeliverRate, 0.0001)
	assert.InDelta(t, 0.1, q.RedliverRate, 0.0001)
}

func TestRMQQueue_UnmarshalJSON_MissingMessageStats(t *testing.T) {
	// RabbitMQ omits "message_stats" entirely for queues with no activity yet.
	data := []byte(`{"name":"idle-queue","vhost":"/","messages":0}`)

	var q models.RMQQueue
	err := json.Unmarshal(data, &q)
	require.NoError(t, err)

	assert.Equal(t, "idle-queue", q.Name)
	assert.Equal(t, 0.0, q.PublishRate)
	assert.Equal(t, 0.0, q.DeliverRate)
	assert.Equal(t, 0.0, q.RedliverRate)
}

func TestRMQQueue_UnmarshalJSON_MalformedJSON(t *testing.T) {
	var q models.RMQQueue
	err := json.Unmarshal([]byte(`{"name":`), &q)
	assert.Error(t, err)
}

func TestRMQQueue_UnmarshalJSON_MalformedRate(t *testing.T) {
	data := []byte(`{"message_stats":{"publish_details":{"rate":"not-a-number"}}}`)
	var q models.RMQQueue
	err := json.Unmarshal(data, &q)
	assert.Error(t, err)
}
