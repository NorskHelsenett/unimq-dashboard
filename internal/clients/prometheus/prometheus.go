package prometheus

import "github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"

type PromClient struct {
	RestClient *rest.RestClient
	Username   string
	Password   string
}
