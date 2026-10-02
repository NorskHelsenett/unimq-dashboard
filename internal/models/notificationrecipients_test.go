package models_test

import (
	"testing"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostRecipient_ToRecipient(t *testing.T) {
	t.Run("webhook recipient", func(t *testing.T) {
		p := &models.PostRecipient{
			Name: "Slack Channel",
			URL:  "https://hooks.slack.com/services/xyz",
			Type: models.RecipientTypeWebhook,
		}

		r, err := p.ToRecipient()
		require.NoError(t, err)
		assert.NotEmpty(t, r.ID)
		assert.Equal(t, "Slack Channel", r.Name)
		assert.Equal(t, "https://hooks.slack.com/services/xyz", r.URL)
		assert.Equal(t, models.RecipientTypeWebhook, r.Type)
	})

	t.Run("email recipient", func(t *testing.T) {
		p := &models.PostRecipient{
			Name:  "Ola Normann",
			Email: "ola.normann@normann.no",
			Type:  models.RecipientTypeEmail,
		}

		r, err := p.ToRecipient()
		require.NoError(t, err)
		assert.Equal(t, "ola.normann@normann.no", r.Email)
		assert.Equal(t, models.RecipientTypeEmail, r.Type)
	})

	t.Run("invalid type is rejected", func(t *testing.T) {
		p := &models.PostRecipient{
			Name: "Bad Recipient",
			Type: "sms",
		}

		r, err := p.ToRecipient()
		require.Error(t, err)
		assert.Nil(t, r)
		assert.Contains(t, err.Error(), "invalid recipient type")
	})
}

func TestParseRecipientType(t *testing.T) {
	cases := []struct {
		in       string
		expected models.RecipientType
	}{
		{"webhook", models.RecipientTypeWebhook},
		{"email", models.RecipientTypeEmail},
		{"sms", models.RecipientTypeUnknown},
		{"", models.RecipientTypeUnknown},
	}

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.expected, models.ParseRecipientType(tc.in))
		})
	}
}

func TestGetReceipientTypes(t *testing.T) {
	types := models.GetReceipientTypes()
	assert.Equal(t, []models.RecipientType{models.RecipientTypeWebhook, models.RecipientTypeEmail}, types)
}

func TestGetRecipientTypesString(t *testing.T) {
	assert.Equal(t, "[webhook, email]", models.GetRecipientTypesString())
}
