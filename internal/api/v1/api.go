package api

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/wneessen/go-mail"
)

type APIService struct {
	Ctx         context.Context
	RMQClient   *rabbitmq.RMQClient
	PromClient  *prometheus.PromClient
	DB          *database.Database
	EmailClient *mail.Client
	EmailConfig *config.EmailConfig
	RMQLimits   *models.Limits
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

func WithEmailConfig(emailConfig *config.EmailConfig) APIServiceOption {
	return func(rc *APIService) error {
		rc.EmailConfig = emailConfig
		return nil
	}
}

func WithChecker(checker *notify.Checker) APIServiceOption {
	return func(rc *APIService) error {
		rc.Checker = checker
		return nil
	}
}

func newAPIServiceConfig() *APIService {
	return &APIService{
		Ctx:         context.Background(),
		RMQClient:   nil,
		PromClient:  nil,
		DB:          nil,
		RMQLimits:   &models.Limits{},
		EmailConfig: nil,
		EmailClient: nil,
	}
}

func NewAPIService(opts ...APIServiceOption) (*APIService, error) {
	rc := newAPIServiceConfig()
	for _, opt := range opts {
		if err := opt(rc); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if rc.EmailConfig != nil && rc.EmailConfig.EmailSMTPHost != "" {

		opts := make([]mail.Option, 0)
		if rc.EmailConfig.EmailSMTPUsername != "" {
			opts = append(opts, mail.WithUsername(rc.EmailConfig.EmailSMTPUsername))
		}
		if rc.EmailConfig.EmailSMTPPassword != "" {
			opts = append(opts, mail.WithPassword(rc.EmailConfig.EmailSMTPPassword))
		}
		if rc.EmailConfig.EmailSMTPPort != 0 {
			opts = append(opts, mail.WithPort(rc.EmailConfig.EmailSMTPPort))
		}
		emailClient, err := mail.NewClient(
			rc.EmailConfig.EmailSMTPHost,
			opts...,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create email client: %w", err)
		}
		rc.EmailClient = emailClient
	}
	return rc, nil
}
