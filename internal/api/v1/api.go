package api

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
)

// TODO: This has become a singleton, and should be refactored to be a proper service with a constructor and options.
type APIService struct {
	Ctx         context.Context
	RMQClient   *rabbitmq.RMQClient
	PromClient  *prometheus.PromClient
	DexClient   *dex.DexClient
	DB          *database.Database
	RMQLimits   *models.Limits
	AdminGroups []string
	Checker     *notify.Checker
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

func WithDexClient(dex *dex.DexClient) APIServiceOption {
	return func(rc *APIService) error {
		rc.DexClient = dex
		return nil
	}
}

func WithDatabase(db *database.Database) APIServiceOption {
	return func(rc *APIService) error {
		rc.DB = db
		return nil
	}
}

func WithChecker(checker *notify.Checker) APIServiceOption {
	return func(rc *APIService) error {
		rc.Checker = checker
		return nil
	}
}

func WithAdminGroups(groups []string) APIServiceOption {
	return func(rc *APIService) error {
		rc.AdminGroups = groups
		return nil
	}
}

func newAPIServiceConfig() *APIService {
	return &APIService{
		Ctx:         context.Background(),
		RMQClient:   nil,
		PromClient:  nil,
		DexClient:   nil,
		DB:          nil,
		RMQLimits:   nil,
		AdminGroups: []string{},
		Checker:     nil,
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
