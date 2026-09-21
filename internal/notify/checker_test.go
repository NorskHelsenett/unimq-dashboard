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
		expectedValue  *float64
		expectedFiring bool
		expectedError  error
	}{
		{
			name:           "channels below threshold",
			ruleType:       models.AlarmTypeChannels,
			threshold:      10,
			metrics:        &models.VhostMetrics{Channels: 5},
			expectedValue:  new(5.0),
			expectedFiring: false,
			expectedError:  nil,
		},
		{
			name:           "channels at threshold triggers",
			ruleType:       models.AlarmTypeChannels,
			threshold:      10,
			metrics:        &models.VhostMetrics{Channels: 10},
			expectedValue:  new(10.0),
			expectedFiring: true,
			expectedError:  nil,
		},
		{
			name:           "channels above threshold triggers",
			ruleType:       models.AlarmTypeChannels,
			threshold:      10,
			metrics:        &models.VhostMetrics{Channels: 15},
			expectedValue:  new(15.0),
			expectedFiring: true,
			expectedError:  nil,
		},
		{
			name:           "connections below threshold",
			ruleType:       models.AlarmTypeConnections,
			threshold:      10,
			metrics:        &models.VhostMetrics{Connections: 3},
			expectedValue:  new(3.0),
			expectedFiring: false,
			expectedError:  nil,
		},
		{
			name:           "connections above threshold triggers",
			ruleType:       models.AlarmTypeConnections,
			threshold:      10,
			metrics:        &models.VhostMetrics{Connections: 42},
			expectedValue:  new(42.0),
			expectedFiring: true,
			expectedError:  nil,
		},
		{
			name:           "queues below threshold",
			ruleType:       models.AlarmTypeQueues,
			threshold:      10,
			metrics:        &models.VhostMetrics{Queues: 2},
			expectedValue:  new(2.0),
			expectedFiring: false,
			expectedError:  nil,
		},
		{
			name:           "queues above threshold triggers",
			ruleType:       models.AlarmTypeQueues,
			threshold:      10,
			metrics:        &models.VhostMetrics{Queues: 11},
			expectedValue:  new(11.0),
			expectedFiring: true,
			expectedError:  nil,
		},
		{
			name:           "nil metrics does not panic and evaluates to zero value",
			ruleType:       models.AlarmTypeChannels,
			threshold:      1,
			metrics:        nil,
			expectedValue:  nil,
			expectedFiring: false,
			expectedError:  notify.ErrNotificationRuleNoMetrics,
		},
		{
			name:           "unknown rule type returns error",
			ruleType:       models.AlarmType("bogus"),
			threshold:      1,
			metrics:        &models.VhostMetrics{Channels: 5},
			expectedValue:  nil,
			expectedFiring: false,
			expectedError:  notify.ErrNotificationRuleUnknownType,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := newRule(tc.ruleType, "", tc.threshold, true)

			result, err := notify.EvaluateMetrics(rule, tc.metrics, nil)

			if tc.expectedError != nil {
				assert.ErrorIsf(t, err, tc.expectedError, "incorrect error for rule type %s", tc.ruleType)
				return
			} else {
				require.NoErrorf(t, err, "unexpected error for rule type %s", tc.ruleType)
			}

			require.NotNilf(t, result, "expected result to not be nil for rule type %s", tc.ruleType)
			assert.Equalf(t, tc.expectedValue, result.Value, "expected value for rule type %s", tc.ruleType)
			assert.Equalf(t, tc.expectedFiring, result.Triggered, "expected firing status for rule type %s", tc.ruleType)
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
		{Name: "test-queue", Messages: 15, Unacked: 7, MessageBytes: 2048, Consumers: 0},
	}

	metrics := &models.VhostMetrics{
		Name:            "test-vhost",
		Connections:     5,
		Channels:        10,
		Queues:          2,
		UnackedMessages: 7,
		ReadyMessages:   15,
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
			t.Parallel()
			rule := newRule(tc.ruleType, queues[0].Name, tc.threshold, true)

			result, err := notify.EvaluateMetrics(rule, metrics, queues)

			require.NoError(t, err)
			require.NotNil(t, result)
			require.NotNil(t, result.Value)
			assert.Equal(t, tc.expectedValue, *result.Value)
			assert.Equal(t, tc.expectedFiring, result.Triggered)
		})
	}
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
			expectedValue:  0,
			expectedFiring: true,
		},
		{
			name:           "messages with consumers does not trigger",
			queue:          models.QueueDetail{Name: "test-queue", Messages: 10, Consumers: 2},
			expectedValue:  2,
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
			t.Parallel()
			rule := newRule(models.AlarmTypeNoConsumer, "test-queue", 0, true)

			result, err := notify.EvaluateMetrics(rule, &models.VhostMetrics{}, []models.QueueDetail{tc.queue})

			require.NoError(t, err)
			require.NotNil(t, result.Value)
			assert.Equalf(t, tc.expectedValue, *result.Value, "incorrect value for queue %s", tc.queue.Name)
			assert.Equal(t, tc.expectedFiring, result.Triggered)
		})
	}
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
