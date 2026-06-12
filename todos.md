- Golangci-lint
- Watchers for mongodb, rabbitmq, prometheus to cover connectivity loss

- Context in Golang for clean shutdown

- api/queues require vhost parameter, should return all and use below example for filtering
- api/queues/{vhost} require queue name parameter
- http.FileServer for static files, should be moved to a separate handler and not used for all routes
- Swagger documentation for API endpoints
- Log entry trimming, should be implemented to prevent log bloat and improve readability
- Retry logic for connecting to RabbitMQ and Prometheus, to handle transient connectivity issues
- Database queries should use the collections parameter instead of hardcoding collection names

---
