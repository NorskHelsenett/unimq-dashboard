package config

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	BaseURL  string `mapstructure:"BASE_URL"`
	BasePort int    `mapstructure:"BASE_PORT"`

	MongoDBHost     string `mapstructure:"MONGODB_HOST"`
	MongoDBPort     int    `mapstructure:"MONGODB_PORT"`
	MongoDBUsername string `mapstructure:"MONGODB_USERNAME"`
	MongoDBPassword string `mapstructure:"MONGODB_PASSWORD"`
	MongoDBDatabase string `mapstructure:"MONGODB_DATABASE"`

	RabbitMQHost            string `mapstructure:"RABBITMQ_HOST"`
	RabbitMQPort            int    `mapstructure:"RABBITMQ_PORT"`
	RabbitMQUsername        string `mapstructure:"RABBITMQ_USERNAME"`
	RabbitMQPassword        string `mapstructure:"RABBITMQ_PASSWORD"`
	RabbitMQChannelLimit    int    `mapstructure:"RABBITMQ_CHANNEL_LIMIT"`
	RabbitMQConnectionLimit int    `mapstructure:"RABBITMQ_CONNECTION_LIMIT"`
	RabbitMQQueueLimit      int    `mapstructure:"RABBITMQ_QUEUE_LIMIT"`

	PrometheusHost string `mapstructure:"PROMETHEUS_HOST"`
	PrometheusPort int    `mapstructure:"PROMETHEUS_PORT"`
}

func NewConfig() *Config {
	c := &Config{
		BaseURL:                 "localhost",
		BasePort:                8080,
		MongoDBHost:             "mongodb://localhost",
		MongoDBPort:             27017,
		MongoDBUsername:         "",
		MongoDBPassword:         "",
		MongoDBDatabase:         "rabbitmq-dashboard",
		RabbitMQHost:            "http://localhost",
		RabbitMQPort:            15672,
		RabbitMQUsername:        "",
		RabbitMQPassword:        "",
		RabbitMQChannelLimit:    1000,
		RabbitMQConnectionLimit: 300,
		RabbitMQQueueLimit:      150,
		PrometheusHost:          "http://localhost",
		PrometheusPort:          9090,
	}
	return c
}

func (c *Config) Load() error {

	viper.AutomaticEnv()
	c.loadEnvironmentVariables()
	_ = c.loadConfigurationFile(".")
	err := viper.Unmarshal(c)
	if err != nil {
		return err
	}

	err = c.validateConfiguration()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) CheckURLs() error {

	rmq := strings.TrimPrefix(c.RabbitMQHost, "http://")
	rmq = strings.TrimPrefix(rmq, "https://")
	_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", rmq, c.RabbitMQPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ URL: %w", err)
	}

	prom := strings.TrimPrefix(c.PrometheusHost, "http://")
	prom = strings.TrimPrefix(prom, "https://")
	_, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", prom, c.PrometheusPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus URL: %w", err)
	}

	mdb := strings.TrimPrefix(c.MongoDBHost, "mongodb://")
	_, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", mdb, c.MongoDBPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB URL: %w", err)
	}

	return nil
}

func (c *Config) loadConfigurationFile(path string) error {

	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("")
	viper.AddConfigPath(".")
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		ok := errors.Is(viper.ConfigFileNotFoundError{}, err)
		if ok {
			return nil
		}
		return err
	}

	err = viper.Unmarshal(&c)

	return err
}

func (c *Config) loadEnvironmentVariables() {

	// Mandatory values
	_ = viper.BindEnv("RABBITMQ_USERNAME")
	_ = viper.BindEnv("RABBITMQ_PASSWORD")
	_ = viper.BindEnv("MONGODB_USERNAME")
	_ = viper.BindEnv("MONGODB_PASSWORD")

	// Defaulted values
	_ = viper.BindEnv("BASE_URL")
	_ = viper.BindEnv("BASE_PORT")
	_ = viper.BindEnv("MONGODB_HOST")
	_ = viper.BindEnv("MONGODB_PORT")
	_ = viper.BindEnv("RABBITMQ_HOST")
	_ = viper.BindEnv("RABBITMQ_PORT")
	_ = viper.BindEnv("PROMETHEUS_PORT")
	_ = viper.BindEnv("PROMETHEUS_HOST")
	_ = viper.BindEnv("RABBITMQ_CHANNEL_LIMIT")
	_ = viper.BindEnv("RABBITMQ_CONNECTION_LIMIT")
	_ = viper.BindEnv("RABBITMQ_QUEUE_LIMIT")
}

func (c *Config) validateConfiguration() error {

	// Checks that any required parameters are present, if any of these are missing going forward is meaingless
	parameterChecks := make(map[string]bool)

	parameterChecks["BASE_URL"] = isPresent(c.BaseURL)
	parameterChecks["BASE_PORT"] = isPresent(c.BasePort)
	parameterChecks["MONGODB_HOST"] = isPresent(c.MongoDBHost)
	parameterChecks["MONGODB_PORT"] = isPresent(c.MongoDBPort)
	parameterChecks["MONGODB_USERNAME"] = isPresent(c.MongoDBUsername)
	parameterChecks["MONGODB_PASSWORD"] = isPresent(c.MongoDBPassword)
	parameterChecks["MONGODB_DATABASE"] = isPresent(c.MongoDBDatabase)
	parameterChecks["RABBITMQ_HOST"] = isPresent(c.RabbitMQHost)
	parameterChecks["RABBITMQ_PORT"] = isPresent(c.RabbitMQPort)
	parameterChecks["RABBITMQ_USERNAME"] = isPresent(c.RabbitMQUsername)
	parameterChecks["RABBITMQ_PASSWORD"] = isPresent(c.RabbitMQPassword)
	parameterChecks["RABBITMQ_CHANNEL_LIMIT"] = isPresent(c.RabbitMQChannelLimit)
	parameterChecks["RABBITMQ_CONNECTION_LIMIT"] = isPresent(c.RabbitMQConnectionLimit)
	parameterChecks["RABBITMQ_QUEUE_LIMIT"] = isPresent(c.RabbitMQQueueLimit)
	parameterChecks["PROMETHEUS_HOST"] = isPresent(c.PrometheusHost)
	parameterChecks["PROMETHEUS_PORT"] = isPresent(c.PrometheusPort)

	errString := checkParameters(parameterChecks)
	if len(errString) > 0 {
		return errors.New(errString)
	}

	return nil
}

func isPresent(value any) bool {
	switch v := value.(type) {
	case string:

		if value == "" {
			return false
		}
		return true
	case int:
		if value == 0 {
			return false
		}
		return true
	default:
		panic(fmt.Sprintf("Unsupported type check of type %v", v))
	}
}

func checkParameters(parameter map[string]bool) string {

	errString := ""

	for parameter, present := range parameter {
		if !present {
			if errString != "" {
				errString += ", "
			}
			errString += parameter + " missing"
		}
	}

	if len(errString) > 0 {
		return errString
	}
	return ""
}
