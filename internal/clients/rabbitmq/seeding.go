package rabbitmq

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/url"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (rmq *RMQClient) Seed(ctx context.Context, vhosts []string) (map[string][]string, error) {

	for _, vhost := range vhosts {
		err := rmq.NewVhost(vhost)
		if err != nil {
			return nil, fmt.Errorf("failed to create vhost %s: %w", vhost, err)
		}
	}

	queueMapping := make(map[string][]string, len(vhosts))

	queuePrefix := "seeded-queue"

	for _, vhost := range vhosts {

		queueCount := rand.IntN(20)
		if queueCount < 5 {
			queueCount = 5
		}

		for count := range queueCount {

			queueName := fmt.Sprintf("%v-%v", queuePrefix, count)
			queueMapping[vhost] = append(queueMapping[vhost], queueName)
			err := rmq.NewQueue(vhost, queueName)
			if err != nil {
				return nil, fmt.Errorf("failed to create queue %s in vhost %s: %w", queueName, vhost, err)
			}
			for i := range 50 {
				payload := fmt.Sprintf(`{"id":%d,"event":"seed","source":"%s/%s"}`, i, vhost, queueName)
				err = rmq.PublishMessage(vhost, queueName, payload)
				if err != nil {
					return nil, fmt.Errorf("failed to publish seed message to %s/%s: %w", vhost, queueName, err)
				}
			}
		}
	}
	return queueMapping, nil
}

func (r *RMQClient) PublishMessage(vhost, queue, payload string) error {
	body := map[string]any{
		"properties":       map[string]any{},
		"routing_key":      queue,
		"payload":          payload,
		"payload_encoding": "string",
	}
	_, err := r.restClient.Post(
		"/exchanges/"+url.PathEscape(vhost)+"/amq.default/publish",
		&body,
		nil,
	)
	return err
}

func (r *RMQClient) NewVhost(name string) error {

	vhost := models.VhostPost{
		Name:             name,
		Description:      "Seeded Vhost instance for RabbitMQ Dashboard",
		Tags:             []string{"seeded"},
		DefaultQueueType: "quorum",
	}
	_, err := r.restClient.Put("/vhosts/"+url.PathEscape(name), &vhost, nil)
	if err != nil {
		return fmt.Errorf("%w. %w", ErrVhostNotFound, err)
	}

	return nil
}

func (r *RMQClient) NewQueue(vhost, name string) error {
	queue := models.QueuePost{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Vhost:      vhost,
		Arguments:  map[string]any{},
	}
	_, err := r.restClient.Put("/queues/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), &queue, nil)
	if err != nil {
		return err
	}
	return nil
}
