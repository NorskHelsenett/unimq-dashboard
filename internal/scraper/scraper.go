package scraper

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type RestClient struct {
	Ctx        context.Context
	RMQClient  *rabbitmq.RMQClient
	PromClient *prometheus.PromClient
	DB         *database.Database
	RMQLimits  *models.Limits
}

type RestClientOption func(*RestClient) error

func WithContext(ctx context.Context) RestClientOption {
	return func(rc *RestClient) error {
		rc.Ctx = ctx
		return nil
	}
}

func WithRabbitMQClient(rmq *rabbitmq.RMQClient) RestClientOption {
	return func(rc *RestClient) error {
		rc.RMQClient = rmq
		return nil
	}
}

func WithPromClient(prom *prometheus.PromClient) RestClientOption {
	return func(rc *RestClient) error {
		rc.PromClient = prom
		return nil
	}
}

func WithDatabase(db *database.Database) RestClientOption {
	return func(rc *RestClient) error {
		rc.DB = db
		return nil
	}
}

func WithRMQLimits(limits *models.Limits) RestClientOption {
	return func(rc *RestClient) error {
		rc.RMQLimits = limits
		return nil
	}
}

func newRestClientConfig() *RestClient {
	return &RestClient{
		Ctx:        context.Background(),
		RMQClient:  nil,
		PromClient: nil,
		DB:         nil,
		RMQLimits:  &models.Limits{},
	}
}

// func NewRestClient(ctx context.Context, baseURL, username, password string, promURL, promAPIVersion string, promPort int, db *database.Database, limits *models.Limits) (*RestClient, error) {
func NewRestClient(opts ...RestClientOption) (*RestClient, error) {
	rc := newRestClientConfig()
	for _, opt := range opts {
		if err := opt(rc); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	return rc, nil
}
