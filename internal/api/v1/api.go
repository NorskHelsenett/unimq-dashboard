package api

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
)

type APIService struct {
	Ctx         context.Context
	DB          *database.Database
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
