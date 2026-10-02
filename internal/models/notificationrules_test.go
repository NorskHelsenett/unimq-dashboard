package models_test

import (
	"testing"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/api/httpsuite"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testRuleName = "rule"

func TestParseAlarmRulePatch(t *testing.T) {
	t.Run("no fields set returns ErrNoFieldsToUpdate", func(t *testing.T) {
		patch := &models.AlarmRulePatch{}
		fields, err := models.ParseAlarmRulePatch(patch)
		require.ErrorIs(t, err, models.ErrNoFieldsToUpdate)
		assert.Nil(t, fields)
	})

	t.Run("all fields set, uses rules.$. mongo prefix", func(t *testing.T) {
		patch := &models.AlarmRulePatch{
			Name:      httpsuite.NewOptional("New Name"),
			Type:      httpsuite.NewOptional(models.AlarmTypeQueueSize),
			QueueName: httpsuite.NewOptional("my-queue"),
			Threshold: httpsuite.NewOptional(42.0),
			Message:   httpsuite.NewOptional("custom message"),
			Enabled:   httpsuite.NewOptional(true),
		}

		fields, err := models.ParseAlarmRulePatch(patch)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{
			"rules.$.name":      "New Name",
			"rules.$.type":      models.AlarmTypeQueueSize,
			"rules.$.queueName": "my-queue",
			"rules.$.threshold": 42.0,
			"rules.$.message":   "custom message",
			"rules.$.enabled":   true,
		}, fields)
	})

	t.Run("partial fields set, only those present are included", func(t *testing.T) {
		patch := &models.AlarmRulePatch{
			Name: httpsuite.NewOptional("Only Name"),
		}

		fields, err := models.ParseAlarmRulePatch(patch)
		require.NoError(t, err)
		assert.Equal(t, map[string]any{
			"rules.$.name": "Only Name",
		}, fields)
	})
}

func TestAlarmRuleCreate_ToAlarmRule(t *testing.T) {
	t.Run("valid alarm type", func(t *testing.T) {
		create := &models.AlarmRuleCreate{
			Name:      "High Queue Size",
			Type:      models.AlarmTypeQueueSize,
			QueueName: "my-queue",
			Threshold: 1000,
			Message:   "custom message",
			Enabled:   true,
		}

		rule, err := create.ToAlarmRule()
		require.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		assert.Equal(t, "High Queue Size", rule.Name)
		assert.Equal(t, models.AlarmTypeQueueSize, rule.Type)
		assert.Equal(t, "my-queue", rule.QueueName)
		assert.Equal(t, 1000.0, rule.Threshold)
		assert.Equal(t, "custom message", rule.Message)
		assert.True(t, rule.Enabled)
		assert.Equal(t, models.AlarmStatusActive, rule.Status)
		assert.Nil(t, rule.LastFired)
		assert.Nil(t, rule.LastValue)
	})

	t.Run("disabled rule starts as inactive", func(t *testing.T) {
		create := &models.AlarmRuleCreate{
			Type:    models.AlarmTypeQueues,
			Enabled: false,
		}

		rule, err := create.ToAlarmRule()
		require.NoError(t, err)
		assert.Equal(t, models.AlarmStatusInactive, rule.Status)
	})

	t.Run("invalid alarm type", func(t *testing.T) {
		create := &models.AlarmRuleCreate{
			Type: "not-a-real-type",
		}

		rule, err := create.ToAlarmRule()
		require.ErrorIs(t, err, models.ErrInvalidAlarmType)
		assert.Nil(t, rule)
	})
}

func TestNewAlarmRule(t *testing.T) {
	rule := models.NewAlarmRule("My Rule", models.AlarmTypeConnections, "", 10, "msg", true)
	assert.NotEmpty(t, rule.ID)
	assert.Equal(t, "My Rule", rule.Name)
	assert.Equal(t, models.AlarmTypeConnections, rule.Type)
	assert.Equal(t, 10.0, rule.Threshold)
	assert.Equal(t, "msg", rule.Message)
	assert.True(t, rule.Enabled)
	assert.Equal(t, models.AlarmStatusActive, rule.Status)
	assert.Nil(t, rule.LastFired)
	assert.Nil(t, rule.LastValue)
}

func TestAlarmRule_IsTriggered(t *testing.T) {
	rule := models.NewAlarmRule(testRuleName, models.AlarmTypeQueueSize, "q", 100, "", true)

	assert.True(t, rule.IsTriggered(100), "value equal to threshold should trigger")
	assert.True(t, rule.IsTriggered(150), "value above threshold should trigger")
	assert.False(t, rule.IsTriggered(99), "value below threshold should not trigger")
}

func TestAlarmRule_IsChanged(t *testing.T) {
	base := func() *models.AlarmRule {
		return models.NewAlarmRule(testRuleName, models.AlarmTypeQueueSize, "q", 100, "msg", true)
	}

	t.Run("identical rules are not changed", func(t *testing.T) {
		a := base()
		b := base()
		// IDs differ (uuid generated) but IsChanged does not compare ID.
		assert.False(t, a.IsChanged(b))
	})

	t.Run("name differs", func(t *testing.T) {
		a, b := base(), base()
		b.Name = "other"
		assert.True(t, a.IsChanged(b))
	})

	t.Run("type differs", func(t *testing.T) {
		a, b := base(), base()
		b.Type = models.AlarmTypeChannels
		assert.True(t, a.IsChanged(b))
	})

	t.Run("queue name differs", func(t *testing.T) {
		a, b := base(), base()
		b.QueueName = "other-queue"
		assert.True(t, a.IsChanged(b))
	})

	t.Run("threshold differs", func(t *testing.T) {
		a, b := base(), base()
		b.Threshold = 200
		assert.True(t, a.IsChanged(b))
	})

	t.Run("message differs", func(t *testing.T) {
		a, b := base(), base()
		b.Message = "other message"
		assert.True(t, a.IsChanged(b))
	})

	t.Run("enabled differs", func(t *testing.T) {
		a, b := base(), base()
		b.Enabled = false
		assert.True(t, a.IsChanged(b))
	})

	t.Run("only ID differs - not considered a change", func(t *testing.T) {
		a, b := base(), base()
		b.ID = "completely-different-id"
		assert.False(t, a.IsChanged(b))
	})

	t.Run("only status/lastfired/lastvalue differ - not considered a change", func(t *testing.T) {
		a, b := base(), base()
		now := time.Now()
		val := 5.0
		b.Status = models.AlarmStatusFiring
		b.LastFired = &now
		b.LastValue = &val
		assert.False(t, a.IsChanged(b))
	})
}

func TestAlarmRule_BuildMessage(t *testing.T) {
	t.Run("custom message takes precedence", func(t *testing.T) {
		rule := &models.AlarmRule{Name: testRuleName, Type: models.AlarmTypeQueueSize, Message: "custom message"}
		assert.Equal(t, "custom message", rule.BuildMessage("/"))
	})

	t.Run("no_consumer uses queue-specific wording", func(t *testing.T) {
		rule := &models.AlarmRule{Name: testRuleName, Type: models.AlarmTypeNoConsumer, QueueName: "my-queue"}
		msg := rule.BuildMessage("/")
		assert.Contains(t, msg, testRuleName)
		assert.Contains(t, msg, "/")
		assert.Contains(t, msg, "my-queue")
		assert.Contains(t, msg, "no consumers")
	})

	t.Run("maintenance has its own wording", func(t *testing.T) {
		rule := &models.AlarmRule{Name: testRuleName, Type: models.AlarmTypeMaintenance}
		msg := rule.BuildMessage("/")
		assert.Contains(t, msg, "maintenance window")
	})

	t.Run("queue-scoped types mention the queue name", func(t *testing.T) {
		for _, typ := range []models.AlarmType{models.AlarmTypeQueueMessages, models.AlarmTypeQueueSize} {
			rule := &models.AlarmRule{Name: testRuleName, Type: typ, QueueName: "my-queue", Threshold: 10}
			msg := rule.BuildMessage("/")
			assert.Contains(t, msg, "my-queue")
			assert.Contains(t, msg, "10")
		}
	})

	t.Run("vhost-scoped types do not mention a queue", func(t *testing.T) {
		for _, typ := range []models.AlarmType{models.AlarmTypeChannels, models.AlarmTypeConnections, models.AlarmTypeQueues, models.AlarmTypeUnacked} {
			rule := &models.AlarmRule{Name: testRuleName, Type: typ, Threshold: 10}
			msg := rule.BuildMessage("/")
			assert.Contains(t, msg, "/")
			assert.NotContains(t, msg, "queue '")
		}
	})

	t.Run("unknown type falls back to default wording", func(t *testing.T) {
		rule := &models.AlarmRule{Name: testRuleName, Type: "not-a-real-type"}
		msg := rule.BuildMessage("/")
		assert.Equal(t, "Alarm 'rule' has been triggered for vhost '/'.", msg)
	})
}
