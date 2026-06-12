package scraper

import (
	"context"
	"fmt"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/prom"
)

type RestClient struct {
	RMQClient  *rabbitmq.RMQClient
	PromClient *prom.PromClient
	DB         *database.Database
	RMQLimits  *models.Limits
}

type RestClientOption func(*RestClient) error

// TODO: Fix parameter soup to functional parameters.
func NewRestClient(ctx context.Context, baseURL, username, password string, promURL, promAPIVersion string, promPort int, db *database.Database, limits *models.Limits) (*RestClient, error) {
	rmq, err := rabbitmq.NewRMQClient(ctx, baseURL, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ client: %w", err)
	}
	return &RestClient{
		RMQClient:  rmq,
		PromClient: prom.NewPromClient(promURL, promAPIVersion, promPort),
		DB:         db,
		RMQLimits:  limits,
	}, nil
}
