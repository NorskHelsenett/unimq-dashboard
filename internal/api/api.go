package api

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type APIService struct {
	Ctx        context.Context
	RMQClient  *rabbitmq.RMQClient
	PromClient *prometheus.PromClient
	DB         *database.Database
	RMQLimits  *models.Limits
}

type APIServiceOption func(*APIService) error

func WithContext(ctx context.Context) APIServiceOption {
	return func(rc *APIService) error {
		rc.Ctx = ctx
		return nil
	}
}

func WithRabbitMQClient(rmq *rabbitmq.RMQClient) APIServiceOption {
	return func(rc *APIService) error {
		rc.RMQClient = rmq
		return nil
	}
}

func WithPromClient(prom *prometheus.PromClient) APIServiceOption {
	return func(rc *APIService) error {
		rc.PromClient = prom
		return nil
	}
}

func WithDatabase(db *database.Database) APIServiceOption {
	return func(rc *APIService) error {
		rc.DB = db
		return nil
	}
}

func WithRMQLimits(limits *models.Limits) APIServiceOption {
	return func(rc *APIService) error {
		rc.RMQLimits = limits
		return nil
	}
}

func newAPIServiceConfig() *APIService {
	return &APIService{
		Ctx:        context.Background(),
		RMQClient:  nil,
		PromClient: nil,
		DB:         nil,
		RMQLimits:  &models.Limits{},
	}
}

func NewAPIService(opts ...APIServiceOption) (*APIService, error) {
	rc := newAPIServiceConfig()
	for _, opt := range opts {
		if err := opt(rc); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}
	return rc, nil
}
