# TODO: This should be moved to issues or discarded based

- Golangci-lint
- Helm chart variables updated with new environment variables and configuration options

## Low priority ToDo's:

- Retry logic for connecting to RabbitMQ and Prometheus, to handle transient connectivity issues
- Watchers for mongodb, rabbitmq, prometheus to cover connectivity loss
- More comprehensive unit tests for all packages, especially the notify and scraper packages
- Checker does unnecessary updates, should only update when there are changes to the alarm state (e.g., from not triggered to triggered, or vice versa)
- api/queues require vhost parameter, should return all and use below example for filtering
- api/queues/{vhost} require queue name parameter
- More seeding data
