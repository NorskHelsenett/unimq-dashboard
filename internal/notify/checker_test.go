package notify_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
)

func newRule(typ models.AlarmType, queueName string, threshold float64, enabled bool) *models.AlarmRule {
	return models.NewAlarmRule("test-rule", typ, queueName, threshold, "", enabled)
}

func TestEvaluateMetrics_DisabledRule(t *testing.T) {
	rule := newRule(models.AlarmTypeChannels, "", 10, false)
	metrics := &models.VhostMetrics{Channels: 20}

	result, err := notify.EvaluateMetrics(rule, metrics, nil)

	require.ErrorIs(t, err, notify.ErrNotificationRuleDisabled)
	assert.Nil(t, result)
}

func TestEvaluateMetrics_MaintenanceRule(t *testing.T) {
	rule := newRule(models.AlarmTypeMaintenance, "", 0, true)

	result, err := notify.EvaluateMetrics(rule, &models.VhostMetrics{}, nil)

	require.ErrorIs(t, err, notify.ErrNotificationRuleInMaintenance)
	assert.Nil(t, result)
}

func TestEvaluateMetrics_MetricBasedTypes(t *testing.T) {
	cases := []struct {
		name           string
		ruleType       models.AlarmType
		threshold      float64
		metrics        *models.VhostMetrics
		expectedValue  float64
		expectedFiring bool
	}{
		{
			name:           "channels below threshold",
			ruleType:       models.AlarmTypeChannels,
			threshold:      10,
			metrics:        &models.VhostMetrics{Channels: 5},
			expectedValue:  5,
			expectedFiring: false,
		},
		{
			name:           "channels at threshold triggers",
			ruleType:       models.AlarmTypeChannels,
			threshold:      10,
			metrics:        &models.VhostMetrics{Channels: 10},
			expectedValue:  10,
			expectedFiring: true,
		},
		{
			name:           "channels above threshold triggers",
			ruleType:       models.AlarmTypeChannels,
			threshold:      10,
			metrics:        &models.VhostMetrics{Channels: 15},
			expectedValue:  15,
			expectedFiring: true,
		},
		{
			name:           "connections below threshold",
			ruleType:       models.AlarmTypeConnections,
			threshold:      10,
			metrics:        &models.VhostMetrics{Connections: 3},
			expectedValue:  3,
			expectedFiring: false,
		},
		{
			name:           "connections above threshold triggers",
			ruleType:       models.AlarmTypeConnections,
			threshold:      10,
			metrics:        &models.VhostMetrics{Connections: 42},
			expectedValue:  42,
			expectedFiring: true,
		},
		{
			name:           "queues below threshold",
			ruleType:       models.AlarmTypeQueues,
			threshold:      10,
			metrics:        &models.VhostMetrics{Queues: 2},
			expectedValue:  2,
			expectedFiring: false,
		},
		{
			name:           "queues above threshold triggers",
			ruleType:       models.AlarmTypeQueues,
			threshold:      10,
			metrics:        &models.VhostMetrics{Queues: 11},
			expectedValue:  11,
			expectedFiring: true,
		},
		{
			name:           "nil metrics does not panic and evaluates to zero value",
			ruleType:       models.AlarmTypeChannels,
			threshold:      1,
			metrics:        nil,
			expectedValue:  0,
			expectedFiring: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := newRule(tc.ruleType, "", tc.threshold, true)

			result, err := notify.EvaluateMetrics(rule, tc.metrics, nil)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Value)
			assert.Equal(t, tc.expectedValue, *result.Value)
			assert.Equal(t, tc.expectedFiring, result.Triggered)
			if tc.expectedFiring {
				assert.Equal(t, models.AlarmStatusFiring, result.NewStatus)
			} else {
				assert.Equal(t, models.AlarmStatusOK, result.NewStatus)
			}
		})
	}
}

func TestEvaluateMetrics_QueueBasedTypes(t *testing.T) {
	queues := []models.QueueDetail{
		{Name: "other-queue", Messages: 999, Unacked: 999, MessageBytes: 999, Consumers: 1},
		{Name: "test-queue", Messages: 15, Unacked: 7, MessageBytes: 2048, Consumers: 0},
	}

	cases := []struct {
		name           string
		ruleType       models.AlarmType
		threshold      float64
		expectedValue  float64
		expectedFiring bool
	}{
		{
			name:           "unacked below threshold",
			ruleType:       models.AlarmTypeUnacked,
			threshold:      100,
			expectedValue:  7,
			expectedFiring: false,
		},
		{
			name:           "unacked above threshold triggers",
			ruleType:       models.AlarmTypeUnacked,
			threshold:      5,
			expectedValue:  7,
			expectedFiring: true,
		},
		{
			name:           "queue_messages below threshold",
			ruleType:       models.AlarmTypeQueueMessages,
			threshold:      100,
			expectedValue:  15,
			expectedFiring: false,
		},
		{
			name:           "queue_messages above threshold triggers",
			ruleType:       models.AlarmTypeQueueMessages,
			threshold:      10,
			expectedValue:  15,
			expectedFiring: true,
		},
		{
			name:           "queue_size below threshold",
			ruleType:       models.AlarmTypeQueueSize,
			threshold:      4096,
			expectedValue:  2048,
			expectedFiring: false,
		},
		{
			name:           "queue_size above threshold triggers",
			ruleType:       models.AlarmTypeQueueSize,
			threshold:      1000,
			expectedValue:  2048,
			expectedFiring: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := newRule(tc.ruleType, "test-queue", tc.threshold, true)

			result, err := notify.EvaluateMetrics(rule, &models.VhostMetrics{}, queues)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Value)
			assert.Equal(t, tc.expectedValue, *result.Value)
			assert.Equal(t, tc.expectedFiring, result.Triggered)
		})
	}
}

func TestEvaluateMetrics_QueueNotFound(t *testing.T) {
	queues := []models.QueueDetail{
		{Name: "other-queue", Messages: 15, Unacked: 7, MessageBytes: 2048, Consumers: 0},
	}
	rule := newRule(models.AlarmTypeQueueMessages, "missing-queue", 1, true)

	result, err := notify.EvaluateMetrics(rule, &models.VhostMetrics{}, queues)

	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.Equal(t, float64(0), *result.Value)
	assert.False(t, result.Triggered)
}

func TestEvaluateMetrics_NoConsumer(t *testing.T) {
	cases := []struct {
		name           string
		queue          models.QueueDetail
		expectedValue  float64
		expectedFiring bool
	}{
		{
			name:           "messages with no consumers triggers",
			queue:          models.QueueDetail{Name: "test-queue", Messages: 10, Consumers: 0},
			expectedValue:  10,
			expectedFiring: true,
		},
		{
			name:           "messages with consumers does not trigger",
			queue:          models.QueueDetail{Name: "test-queue", Messages: 10, Consumers: 2},
			expectedValue:  10,
			expectedFiring: false,
		},
		{
			name:           "no messages and no consumers does not trigger",
			queue:          models.QueueDetail{Name: "test-queue", Messages: 0, Consumers: 0},
			expectedValue:  0,
			expectedFiring: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := newRule(models.AlarmTypeNoConsumer, "test-queue", 0, true)

			result, err := notify.EvaluateMetrics(rule, &models.VhostMetrics{}, []models.QueueDetail{tc.queue})

			require.NoError(t, err)
			require.NotNil(t, result.Value)
			assert.Equal(t, tc.expectedValue, *result.Value)
			assert.Equal(t, tc.expectedFiring, result.Triggered)
		})
	}
}

func TestEvaluateMetrics_UnknownType(t *testing.T) {
	rule := newRule(models.AlarmType("bogus"), "", 0, true)

	result, err := notify.EvaluateMetrics(rule, &models.VhostMetrics{}, nil)

	require.NoError(t, err)
	require.NotNil(t, result.Value)
	assert.Equal(t, float64(0), *result.Value)
	assert.False(t, result.Triggered)
}

func TestEvaluateRule_DisabledRule(t *testing.T) {
	rule := newRule(models.AlarmTypeChannels, "", 1, false)

	entry, err := notify.EvaluateRule(rule, models.AlarmStatusFiring, 5)

	require.ErrorIs(t, err, notify.ErrNotificationRuleDisabled)
	assert.Nil(t, entry)
}

func TestEvaluateRule_TransitionsToFiring(t *testing.T) {
	rule := newRule(models.AlarmTypeChannels, "", 1, true)
	rule.Status = models.AlarmStatusOK

	entry, err := notify.EvaluateRule(rule, models.AlarmStatusFiring, 5)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, rule.ID, entry.AlarmID)
	require.Len(t, entry.Entries, 1)
	assert.Equal(t, models.LogEventFired, entry.Entries[0].Event)
	require.NotNil(t, entry.Entries[0].Value)
	assert.Equal(t, float64(5), *entry.Entries[0].Value)
}

func TestEvaluateRule_TransitionsToResolved(t *testing.T) {
	rule := newRule(models.AlarmTypeChannels, "", 1, true)
	rule.Status = models.AlarmStatusFiring

	entry, err := notify.EvaluateRule(rule, models.AlarmStatusOK, 0)

	require.NoError(t, err)
	require.NotNil(t, entry)
	assert.Equal(t, rule.ID, entry.AlarmID)
	require.Len(t, entry.Entries, 1)
	assert.Equal(t, models.LogEventResolved, entry.Entries[0].Event)
}

func TestEvaluateRule_NoChange(t *testing.T) {
	cases := []struct {
		name          string
		currentStatus models.AlarmStatus
		newStatus     models.AlarmStatus
	}{
		{name: "ok stays ok", currentStatus: models.AlarmStatusOK, newStatus: models.AlarmStatusOK},
		{name: "firing stays firing", currentStatus: models.AlarmStatusFiring, newStatus: models.AlarmStatusFiring},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := newRule(models.AlarmTypeChannels, "", 1, true)
			rule.Status = tc.currentStatus

			entry, err := notify.EvaluateRule(rule, tc.newStatus, 1)

			require.ErrorIs(t, err, notify.ErrNotificationRuleNoChange)
			assert.Nil(t, entry)
		})
	}
}
