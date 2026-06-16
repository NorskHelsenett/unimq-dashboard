# UniMQ Dashboard

Et overvåkingsdashboard og varslingsside for UniMQ. Viser metrikker per vhost, kø-detaljer med historiske grafer, vedlikeholdsplanlegging og webhook-baserte alarmer.

---

## Kjøre dashboardet med ny Dex authentication

For å få tilgang til dashboardet er det nå nødvendig og også spinne opp en container som kjører dex.

Slik det er satt opp nå kjører man denne containeren, Go applikasjonen og Vite dev server samtidig ved å kjøre:

```bash
cd web-src
npm run dev:all
```

---

## Funksjonalitet

- **Oversikt** — Connections, channels, køer og unacked meldinger per vhost, med grenser og fargeindikator
- **Kø-tabell** — Meldinger, størrelse, consumers, publish/deliver-rate og sparkline-grafer per kø
- **Kø-detaljer** — Historisk graf med Prometheus-data (5m, 1h, 6h, 24h, 7d)
- **Cluster Resources** — Minne- og disk-gauger for clusteret, meldingsdata for valgt vhost
- **Alarmer** — Konfigurerbare alarmer for connections, channels, køer, unacked, meldinger i kø, kø-størrelse, ingen consumer og vedlikehold
- **Webhook-varsling** — Alarmer kan sende varsler til Slack, Microsoft Teams eller andre tjenester via innkommende webhooks ved alarmutløsning. Meldinger sendes som HTTP POST med JSON-payload:
  `json
{ "text": "[UniMQ] Alarm: <navn> — <vhost>\n\n<beskrivelse>" }
`
- **Vedlikehold** — Synliggjøre vedlikeholdsvindu og eventuelle endringer i forbindelse med oppdatering av RabbitMQ og OS. Leveranseteam integrasjon skal kunne publisere endringer.

---

## Prosjektstruktur

```
unimq-dashboard/
├── cmd/
│   ├── unimq/
│   │   └── main.go                # Main application that starts the API server and background tasks
│   └── generator/
│       └── main.go                # Generates mock data for testing and development (optional)
├── internal/
│   ├── api/                       # API handlers and business logic for the HTTP server
│   ├── clients/
│   │   ├── prometheus/            # Prometheus HTTP client
│   │   ├── rabbitmq/              # RabbitMQ HTTP client (Management API)
│   │   └── rest/                  # Generic REST client for making HTTP requests
│   │       └── httpauthproviders/ # Auth providers for http client (basic auth, bearer token, etc.)
│   ├── config/                    # Configuration loading from environment variables
│   ├── database/                  # Database quries for data persistence (alarms, maintenance, etc.)
│   ├── logger/                    # Logger setup and utilities
│   ├── models/                    # Data models for everything used in the application (alarms, maintenance, API responses, etc.)
│   ├── notificationhelper/        # Helper functions for formatting and sending notifications (e.g., to Slack, Teams)
│   ├── notify/                    # Checker for alarms and sending notifications
│   ├── routes/                    # HTTP route handlers for the API endpoints
│   │   └── httpsuite/             # generic http handler for responses, errors
│   └── templating/                # Helper functions for rendering HTML templates
├── scripts/
├── unimq/                         # helm charts for deployment of UniMQ
├── volumes-for-compose/           # Docker files and volumes for local development environment
├── web/
│   ├── static/
│   │   ├── style.css              # All CSS for dashboardet
│   │   └── logo.png               # UniMQ-logo
│   └── templates/
│       ├── index.html             # Oversiktsside med limits, kø-tabell og cluster-gauger
│       ├── queue.html             # Kø-detaljside med historisk Prometheus-graf
│       ├── maintenance.html       # Vedlikeholdsside for brukere
│       ├── maintenance_admin.html # Admin-side for å legge til/endre vedlikehold
│       ├── notifications.html     # Alarmkonfigurasjon og webhook-mottakere
│       └── notification_rule.html # Detaljside for én alarm (rediger, test, sist utløst)
├── web-src/
├── docker-bake.hcl
└── docker-compose.yaml
```

---

## Development environment - Backend

### Prerequisites

- [Go](https://go.dev/) 1.26.4 or newer
- Docker and Docker Compose for running RabbitMQ, Prometheus, mongodb, and Dex locally

### 1. Clone the repo

```bash
git clone https://github.com/NorskHelsenett/unimq-dashboard.git
cd unimq-dashboard
```

### 2. Copy the .env.example file and update the environment variables as needed

```bash
cp .env.example .env
```

### 3. Start the local development environment with Docker Compose

```bash
docker-compose up -d
```

### 4. Start the dashboard application

```bash
go run ./cmd/unimq/main.go
```

## Build with docker-bake

```bash
docker buildx bake all
```
