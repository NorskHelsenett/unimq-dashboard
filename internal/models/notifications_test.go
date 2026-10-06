package models_test

import (
	"testing"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAlarmTypes(t *testing.T) {
	types := models.GetAlarmTypes()
	assert.Len(t, types, 8)
	assert.Contains(t, types, models.AlarmTypeQueueSize)
	assert.Contains(t, types, models.AlarmTypeMaintenance)
}

func TestGetQueueAlarmTypes(t *testing.T) {
	types := models.GetQueueAlarmTypes()
	assert.ElementsMatch(t, []models.AlarmType{
		models.AlarmTypeQueueMessages,
		models.AlarmTypeQueueSize,
		models.AlarmTypeNoConsumer,
		models.AlarmTypeUnacked,
	}, types)
}

func TestGetVhostAlarmTypes(t *testing.T) {
	types := models.GetVhostAlarmTypes()
	assert.ElementsMatch(t, []models.AlarmType{
		models.AlarmTypeChannels,
		models.AlarmTypeConnections,
		models.AlarmTypeQueues,
	}, types)
}

func TestIsValidAlarmType(t *testing.T) {
	assert.True(t, models.IsValidAlarmType("queue_size"))
	assert.True(t, models.IsValidAlarmType("maintenance"))
	assert.False(t, models.IsValidAlarmType("not-a-real-type"))
	assert.False(t, models.IsValidAlarmType(""))
}

func TestConvertToAlarmType(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		typ, err := models.ConvertToAlarmType("queue_size")
		require.NoError(t, err)
		assert.Equal(t, models.AlarmTypeQueueSize, typ)
	})

	t.Run("invalid", func(t *testing.T) {
		typ, err := models.ConvertToAlarmType("not-a-real-type")
		require.Error(t, err)
		assert.Equal(t, models.AlarmType(""), typ)
		assert.Contains(t, err.Error(), "invalid alarm type")
	})
}

func TestNewVhostNotification(t *testing.T) {
	vn := models.NewVhostNotification("/")
	assert.Equal(t, "/", vn.Name)
	assert.Empty(t, vn.Recipients)
	assert.Empty(t, vn.Rules)
	assert.False(t, vn.Notified)
}

func TestVhostNotification_WebhookURLs(t *testing.T) {
	vn := models.NewVhostNotification("/")
	vn.Recipients = []*models.Recipient{
		{Type: models.RecipientTypeWebhook, URL: "https://example.com/hook1"},
		{Type: models.RecipientTypeWebhook, URL: ""}, // empty URL should be skipped
		{Type: models.RecipientTypeEmail, Email: "a@example.com"},
		{Type: models.RecipientTypeWebhook, URL: "https://example.com/hook2"},
	}

	urls := vn.WebhookURLs()
	assert.Equal(t, []string{"https://example.com/hook1", "https://example.com/hook2"}, urls)
}

func TestVhostNotification_WebhookURLs_Empty(t *testing.T) {
	vn := models.NewVhostNotification("/")
	urls := vn.WebhookURLs()
	assert.NotNil(t, urls, "should return an empty slice, not nil")
	assert.Empty(t, urls)
}

func TestVhostNotification_EmailRecipients(t *testing.T) {
	vn := models.NewVhostNotification("/")
	vn.Recipients = []*models.Recipient{
		{Type: models.RecipientTypeEmail, Email: "a@example.com"},
		{Type: models.RecipientTypeEmail, Email: ""}, // empty email should be skipped
		{Type: models.RecipientTypeWebhook, URL: "https://example.com/hook"},
		{Type: models.RecipientTypeEmail, Email: "b@example.com"},
	}

	emails := vn.EmailRecipients()
	assert.Equal(t, []string{"a@example.com", "b@example.com"}, emails)
}

func TestVhostNotification_EmailRecipients_Empty(t *testing.T) {
	vn := models.NewVhostNotification("/")
	emails := vn.EmailRecipients()
	assert.NotNil(t, emails, "should return an empty slice, not nil")
	assert.Empty(t, emails)
}
