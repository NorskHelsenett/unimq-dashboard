package routes

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	api "github.com/sisneve/rabbitmq-dashboard/internal/api/v1"
	"github.com/sisneve/rabbitmq-dashboard/internal/api/v1/rmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/dex"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	_ "github.com/sisneve/rabbitmq-dashboard/internal/docs"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
)

func SetupRoutes(ctx context.Context, config *config.Config, db *database.Database, rmqclient *rabbitmq.RMQClient, checker *notify.Checker) (chi.Router, error) {

	dex, err := dex.NewDexClient(ctx, config.OIDC)
	if err != nil {
		return nil, fmt.Errorf("failed to create Dex client: %w", err)
	}

	apiservice, err := api.NewAPIService(
		api.WithContext(ctx),
		api.WithDatabase(db),
		api.WithChecker(checker),
		api.WithAdminGroups(config.AdminGroups),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create API service: %w", err)
	}

	rmqHandler := rmq.NewRMQHandler(rmqclient, config.AdminGroups)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			SetupUnprotectedRoutes(r, apiservice, dex)
			SetupAuthenticationRoutes(r, dex)

			r.Group(func(r chi.Router) {
				r.Use(dex.Authorization())
				SetupProtectedRoutes(r, apiservice, rmqHandler)
			})
		})
	})

	routeCount := 0
	// Logs every route implicitly or explicitly defined above with its method, path, and number of middlewares.
	err = chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		routeCount++
		slog.InfoContext(ctx, "route info", "method", method, "route", route, "middlewares", len(middlewares))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk routes: %w", err)
	}

	slog.InfoContext(ctx, "total routes registered", "count", routeCount)

	return r, nil

}
