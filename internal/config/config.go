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
	BaseURL          string `mapstructure:"BASE_URL"`
	BasePort         int    `mapstructure:"BASE_PORT"`
	RabbitMQURL      string `mapstructure:"RABBITMQ_URL"`
	RabbitMQPort     int    `mapstructure:"RABBITMQ_PORT"`
	RabbitMQUsername string `mapstructure:"RABBITMQ_USERNAME"`
	RabbitMQPassword string `mapstructure:"RABBITMQ_PASSWORD"`
	PrometheusURL    string `mapstructure:"PROMETHEUS_URL"`
	PrometheusPort   int    `mapstructure:"PROMETHEUS_PORT"`
}

func NewConfig() *Config {
	c := &Config{
		BaseURL:          "localhost",
		BasePort:         8080,
		RabbitMQURL:      "http://localhost",
		RabbitMQPort:     15672,
		RabbitMQUsername: "",
		RabbitMQPassword: "",
		PrometheusURL:    "http://localhost",
		PrometheusPort:   9090,
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

	rmq := strings.TrimPrefix(c.RabbitMQURL, "http://")
	rmq = strings.TrimPrefix(rmq, "https://")

	_, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", rmq, c.RabbitMQPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ URL: %w", err)
	}

	prom := strings.TrimPrefix(c.PrometheusURL, "http://")
	prom = strings.TrimPrefix(prom, "https://")

	_, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", prom, c.PrometheusPort), 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to Prometheus URL: %w", err)
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

	// Defaulted values
	_ = viper.BindEnv("BASE_URL")
	_ = viper.BindEnv("BASE_PORT")
	_ = viper.BindEnv("RABBITMQ_URL")
	_ = viper.BindEnv("RABBITMQ_PORT")
	_ = viper.BindEnv("PROMETHEUS_PORT")
	_ = viper.BindEnv("PROMETHEUS_URL")
}

func (c *Config) validateConfiguration() error {

	// Checks that any required parameters are present, if any of these are missing going forward is meaingless
	parameterChecks := make(map[string]bool)

	parameterChecks["URL"] = isPresent(c.BaseURL)
	parameterChecks["PORT"] = isPresent(c.BasePort)
	parameterChecks["RABBITMQ_URL"] = isPresent(c.RabbitMQURL)
	parameterChecks["RABBITMQ_PORT"] = isPresent(c.RabbitMQPort)
	parameterChecks["RABBITMQ_USERNAME"] = isPresent(c.RabbitMQUsername)
	parameterChecks["RABBITMQ_PASSWORD"] = isPresent(c.RabbitMQPassword)
	parameterChecks["PROMETHEUS_URL"] = isPresent(c.PrometheusURL)
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
