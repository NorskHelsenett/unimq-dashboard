# RabbitMQ Dashboard

Go-nettside som viser vhost-metrikker fra RabbitMQ Management API med embedded Grafana-dashboard.

## Port Check (kjør ved session-start)

```bash
lsof -i :5672 -i :15672 -i :15692 -i :9090 -i :3000 -i :8080 | grep LISTEN
```

| Port  | Tjeneste                    |
|-------|-----------------------------|
| 5672  | RabbitMQ AMQP               |
| 15672 | RabbitMQ Management API     |
| 15692 | RabbitMQ Prometheus metrics |
| 9090  | Prometheus                  |
| 3000  | Grafana                     |
| 8080  | Go-app (denne appen)        |

Start tjenester hvis de mangler:
```bash
brew services start rabbitmq
brew services start prometheus
brew services start grafana
```

## Prosjektstruktur

```
rabbitmq-dashboard/
├── cmd/main.go              # Entrypoint
├── internal/
│   ├── scraper/scraper.go   # Henter data fra RabbitMQ Management API
│   └── api/api.go           # HTTP-endepunkter
├── web/
│   ├── static/              # CSS/JS
│   └── templates/           # HTML-templates
└── go.mod
```

## RabbitMQ Management API

Base URL: `http://localhost:15672/api/` (guest/guest)

Relevante endepunkter:
- `/api/vhosts` — liste over vhoster
- `/api/connections` — alle connections (filtrer på vhost)
- `/api/channels` — alle channels (filtrer på vhost)
- `/api/queues/{vhost}` — køer per vhost
- `/api/vhosts/{vhost}` — statistikk inkl. unacked

## Ressursgrenser per vhost

| Metrikk      | Grense    |
|--------------|-----------|
| Connections  | 300       |
| Queues       | 150       |
| Messages/kø  | 10 000    |
| Størrelse/kø | 10 GiB    |

## Grafana

- UI: http://localhost:3000 (admin/admin)
- Dashboard UID: `rabbitmq-queues-overview`
- Embed URL: `http://localhost:3000/d/rabbitmq-queues-overview?orgId=1&kiosk`
- Provisioning: `/opt/homebrew/opt/grafana/share/grafana/conf/provisioning/`

## Kjør appen

```bash
go run cmd/main.go
```
