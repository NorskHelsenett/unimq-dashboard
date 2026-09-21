package notify

import (
	"fmt"
	"slices"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type evaluationResult struct {
	Triggered bool
	Value     *float64
	NewStatus models.AlarmStatus
}

// EvaluateMetrics evaluates a single alarm rule against the current metrics and returns whether it is triggered, the current value, and the new status.
func EvaluateMetrics(rule *models.AlarmRule, metrics *models.VhostMetrics, queues []models.QueueDetail) (*evaluationResult, error) {

	if !rule.Enabled {
		return nil, ErrNotificationRuleDisabled
	}
	if rule.Type == models.AlarmTypeMaintenance {
		return nil, ErrNotificationRuleInMaintenance
	}

	triggered := false
	var val *float64
	var err error

	switch {
	case slices.Contains(models.GetQueueAlarmTypes(), rule.Type):
		triggered, val, err = evaluateQueueMetrics(rule, queues)
		if err != nil {
			return nil, err
		}
	case slices.Contains(models.GetVhostAlarmTypes(), rule.Type):
		triggered, val, err = evaluateVhostMetrics(rule, metrics)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown rule type: %s, %w", rule.Type, ErrNotificationRuleUnknownType)
	}

	newStatus := models.AlarmStatusOK
	if triggered {
		newStatus = models.AlarmStatusFiring
	}

	output := &evaluationResult{
		Triggered: triggered,
		Value:     val,
		NewStatus: newStatus,
	}

	return output, nil
}

func evaluateVhostMetrics(rule *models.AlarmRule, metrics *models.VhostMetrics) (bool, *float64, error) {
	if metrics == nil {
		return false, nil, fmt.Errorf("metrics are nil, %w", ErrNotificationRuleNoMetrics)
	}

	var v *float64
	switch rule.Type {
	case models.AlarmTypeChannels:
		v = new(float64(metrics.Channels))
	case models.AlarmTypeConnections:
		v = new(float64(metrics.Connections))
	case models.AlarmTypeQueues:
		v = new(float64(metrics.Queues))
	default:
		return false, nil, fmt.Errorf("unknown rule type: %s, %w", rule.Type, ErrNotificationRuleUnknownType)
	}

	if v == nil {
		return false, nil, fmt.Errorf("metric is nil, %w", ErrNotificationRuleEvaluationFailed)
	}

	return *v >= rule.Threshold, v, nil
}

func evaluateQueueMetrics(rule *models.AlarmRule, queues []models.QueueDetail) (bool, *float64, error) {
	if len(queues) == 0 {
		return false, nil, fmt.Errorf("queue metrics are nil, %w", ErrNotificationRuleNoqueueMetrics)
	}

	var v *float64
	queueFound := false
	for _, q := range queues {
		if q.Name == rule.QueueName {
			queueFound = true
			switch rule.Type {
			case models.AlarmTypeUnacked:
				v = new(float64(q.Unacked))
			case models.AlarmTypeQueueMessages:
				v = new(float64(q.Messages))
			case models.AlarmTypeQueueSize:
				v = new(float64(q.MessageBytes))
			case models.AlarmTypeNoConsumer:
				v = new(float64(q.Messages))
				return q.Messages > 0 && q.Consumers == 0, new(float64(q.Consumers)), nil
			default:
				return false, nil, fmt.Errorf("unknown rule type: %s, %w", rule.Type, ErrNotificationRuleUnknownType)
			}
			break
		}
	}

	if !queueFound {
		return false, nil, fmt.Errorf("%w: %s", ErrNotificationRuleQueueNotFound, rule.QueueName)
	}

	if v == nil {
		return false, nil, fmt.Errorf("metric is nil, %w", ErrNotificationRuleEvaluationFailed)
	}

	return *v >= rule.Threshold, v, nil
}

// EvaluateRule checks if the status of the rule has changed and returns an AlarmEntry if it has.
func EvaluateRule(rule *models.AlarmRule, newStatus models.AlarmStatus, newValue float64) (*models.AlarmEntry, error) {

	if !rule.Enabled {
		return nil, ErrNotificationRuleDisabled
	}

	if rule.Status != models.AlarmStatusFiring && newStatus == models.AlarmStatusFiring {
		entry := models.NewLogEntry(models.LogEventFired, &newValue, rule.Threshold, rule.Type)
		alarm := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{entry},
		}

		return &alarm, nil

	} else if rule.Status == models.AlarmStatusFiring && newStatus == models.AlarmStatusOK {
		entry := models.NewLogEntry(models.LogEventResolved, &newValue, rule.Threshold, rule.Type)
		alarm := models.AlarmEntry{
			AlarmID: rule.ID,
			Entries: []models.LogEntry{entry},
		}
		return &alarm, nil
	}

	return nil, ErrNotificationRuleNoChange
}
