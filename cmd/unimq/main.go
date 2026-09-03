package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rabbitmq"
	"github.com/sisneve/rabbitmq-dashboard/internal/config"
	"github.com/sisneve/rabbitmq-dashboard/internal/database"
	"github.com/sisneve/rabbitmq-dashboard/internal/logger"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
	"github.com/sisneve/rabbitmq-dashboard/internal/notify"
	"github.com/sisneve/rabbitmq-dashboard/internal/routes"
)

//	@title			RabbitMQ Dashboard API
//	@version		1.0
//	@description	API for RabbitMQ Dashboard application.

//	@contact.name	Norsk helsenett SF
//	@contact.url	https://github.com/NorskHelsenett/unimq-dashboard

//	@host					localhost:8080
//	@basePath				/api
//	@securityDefinitions	bearer					Authorization
//	@in						header	Authorization	"Bearer {token}"
//	@name					Authorization

func main() {

	logger.SetupLogger()

	config := config.NewConfig()
	if err := config.Load(); err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	logger.UpdateLogLevel(slog.Level(config.LogLevel))

	err := config.CheckURLs()
	if err != nil {
		slog.Error("failed to validate URLs", "error", err)
		return
	}

	db, err := database.NewDatabase(
		database.WithDatabase(config.MongoDBDatabase),
		database.WithHost(config.MongoDBHost),
		database.WithPort(config.MongoDBPort),
		database.WithUsername(config.MongoDBUsername),
		database.WithPassword(config.MongoDBPassword),
	)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	limits := &models.Limits{
		MaxChannels:    config.RabbitMQChannelLimit,
		MaxConnections: config.RabbitMQConnectionLimit,
		MaxQueues:      config.RabbitMQQueueLimit,
	}

	rmq, err := rabbitmq.NewRMQClient(
		rabbitmq.WithRMQHost(config.RabbitMQHost),
		rabbitmq.WithRMQPort(config.RabbitMQPort),
		rabbitmq.WithRMQUsername(config.RabbitMQUsername),
		rabbitmq.WithRMQPassword(config.RabbitMQPassword),
		rabbitmq.WithRMQContext(ctx),
		rabbitmq.WithRMQLimits(limits),
	)
	if err != nil {
		slog.Error("failed to create RabbitMQ client", "error", err)
		return
	}

	checker := notify.NewChecker(
		notify.WithDB(db),
		notify.WithRMQClient(rmq),
		notify.WithInterval(60*time.Second),
		notify.WithContext(ctx),
	)

	routes, err := routes.SetupRoutes(ctx, config, db, rmq, checker)
	if err != nil {
		slog.Error("failed to set up routes", "error", err)
		return
	}

	slog.Info("starting RabbitMQ Dashboard", "URL", config.BaseURL, "port", config.BasePort)
	slog.Info("Swagger documentation available at", "URL", fmt.Sprintf("http://%v:%d/api/swagger/index.html", config.BaseURL, config.BasePort))
	wg := &sync.WaitGroup{}

	server := &http.Server{
		Addr:         fmt.Sprintf("%v:%d", config.BaseURL, config.BasePort),
		Handler:      routes,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	wg.Add(1)
	wg.Go(func() {
		defer wg.Done()

		err = server.ListenAndServe()
		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("http server closed")
				return
			}
			slog.Error("failed to start server", "error", err)
			return
		}
	})

	wg.Add(1)
	checker.StartChecker(wg)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
	case <-ctx.Done():
	}

	slog.Info("shutting down server...")
	cancel()
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer timeoutCancel()
	err = server.Shutdown(timeoutCtx)
	if err != nil {
		slog.Error("failed to shut down server gracefully", "error", err)
		err = server.Close()
		if err != nil {
			slog.Error("forced shutdown of server", "error", err)
		}
	} else {
		wg.Wait()
		slog.Info("server stopped gracefully, good bye :)")
	}

}
