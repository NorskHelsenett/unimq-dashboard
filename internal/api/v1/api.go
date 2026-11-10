package api

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/mongodb"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/sisneve/rabbitmq-dashboard/internal/services/storeservice"
)

type APIService struct {
	Ctx          context.Context
	DB           *mongodb.Database
	RMQClient    rabbitmq.RMQClientInterface
	store        database.Store
	storeService *storeservice.StoreService
	AdminGroups  []string
	Checker      *notify.Checker
}

type APIServiceOption func(*APIService) error

func WithContext(ctx context.Context) APIServiceOption {
	return func(rc *APIService) error {
		rc.Ctx = ctx
		return nil
	}
}

func WithDatabase(db *mongodb.Database) APIServiceOption {
	return func(rc *APIService) error {
		rc.DB = db
		rc.store = db
		rc.storeService, _ = storeservice.NewStoreService(db)
		return nil
	}
}

func WithRMQClient(client rabbitmq.RMQClientInterface) APIServiceOption {
	return func(rc *APIService) error {
		rc.RMQClient = client
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
		DB:          nil,
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
