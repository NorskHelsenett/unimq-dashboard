package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sisneve/rabbitmq-dashboard/internal/api"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/prometheus"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	_ "github.com/sisneve/rabbitmq-dashboard/internal/docs"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func SetupRoutes(ctx context.Context, config *config.Config, db *database.Database, rmq *rabbitmq.RMQClient) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(middleware.RequestID)

	prom, err := prometheus.NewPromClient(config.PrometheusHost, "v1", "", "", config.PrometheusPort)
	if err != nil {
		return nil, fmt.Errorf("failed to create Prometheus client: %w", err)
	}

	limits := &models.Limits{
		MaxChannels:    config.RabbitMQChannelLimit,
		MaxConnections: config.RabbitMQConnectionLimit,
		MaxQueues:      config.RabbitMQQueueLimit,
	}

	apiservice, err := api.NewAPIService(
		api.WithContext(ctx),
		api.WithRabbitMQClient(rmq),
		api.WithPromClient(prom),
		api.WithDatabase(db),
		api.WithRMQLimits(limits),
	)

	r.Route("/api", func(r chi.Router) {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
		SetupV1Routes(r, apiservice)
	})

	// Logs every route implicitly or explicitly defined above with its method, path, and number of middlewares.
	chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		slog.Info("route info", "method", method, "route", route, "middlewares", len(r.Middlewares()))
		return nil
	})

	return r, nil

}
