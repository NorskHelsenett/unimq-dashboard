package notify_test

import (
	"testing"

	"github.com/go-jose/go-jose/v4/testutils/require"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
)

type MetricsTestCase struct {
	Name           string
	Rule           *models.AlarmRule
	Metric         *models.VhostMetrics
	Queues         []models.QueueDetail
	ExpectedResult bool
}

func TestEvaluateMetrics(t *testing.T) {

	cases := []MetricsTestCase{
		{
			Name: "Test case 1: Queue size exceeds threshold",
			Rule: models.NewAlarmRule("queue length", models.AlarmTypeQueueSize, "test-queue", 10, "Queue length exceeded threshold", true),
			Metric: &models.VhostMetrics{
				Name:        "/",
				Connections: 0,
				Channels:    0,
				Queues:      15,
				Unacked:     0,
			},
			Queues: []models.QueueDetail{
				{
					Name:     "test-queue",
					Messages: 15,
				},
			},
			ExpectedResult: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			result, err := notify.EvaluateMetrics(tc.Rule, tc.Metric, tc.Queues)
			require.NoError(t, err, "EvaluateMetrics returned an error")
			if result.Triggered != tc.ExpectedResult {
				t.Errorf("Expected %v, but got %v", tc.ExpectedResult, result)
			}
		})
	}

}

type RulesTestCase struct {
	Name           string
	rule           *models.AlarmRule
	metrics        *models.VhostMetrics
	queues         []models.QueueDetail
	ExpectedResult bool
}

func TestEvaluateRules(t *testing.T) {
}
