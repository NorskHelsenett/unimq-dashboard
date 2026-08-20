package rabbitmq

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

func (rmq *RMQClient) Seed(ctx context.Context) error {

	vhosts := []string{"unimq", "unimq-test"}
	for _, vhost := range vhosts {
		err := rmq.NewVhost(vhost)
		if err != nil {
			return fmt.Errorf("failed to create vhost %s: %w", vhost, err)
		}
	}

	queues := []string{"unimq-queue", "unimq-test-queue"}
	for _, queue := range queues {
		for _, vhost := range vhosts {
			err := rmq.NewQueue(vhost, queue)
			if err != nil {
				return fmt.Errorf("failed to create queue %s in vhost %s: %w", queue, vhost, err)
			}
			for i := range 25 {
				payload := fmt.Sprintf(`{"id":%d,"event":"seed","source":"%s/%s"}`, i, vhost, queue)
				err = rmq.PublishMessage(vhost, queue, payload)
				if err != nil {
					return fmt.Errorf("failed to publish seed message to %s/%s: %w", vhost, queue, err)
				}
			}
		}
	}
	return nil
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
