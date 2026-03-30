# UniMQ Dashboard

Et overvåkingsdashboard og varslingsside for UniMQ. Viser metrikker per vhost, kø-detaljer med historiske grafer, vedlikeholdsplanlegging og webhook-baserte alarmer.

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
│   └── main.go                    # HTTP-server, ruter og side-handlere
├── internal/
│   ├── scraper/
│   │   └── scraper.go             # Henter metrikker fra RabbitMQ Management API
│   ├── prom/
│   │   └── prom.go                # Henter historiske data fra Prometheus
│   ├── maintenance/
│   │   └── store.go               # Lagrer og leser vedlikeholdsoppføringer (JSON)
│   └── notify/
│       ├── store.go               # Alarm- og mottakerdata, webhook-utsending
│       └── checker.go             # Bakgrunnssjekk av alarmer hvert 60. sekund
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
├── data/
│   ├── maintenance.json           # Persistent lagring av vedlikeholdsoppføringer
│   └── notifications.json        # Persistent lagring av alarmer og mottakere
└── go.mod
```

---

## Kom i gang

### Forutsetninger

- [Go](https://go.dev/) 1.21 eller nyere
- RabbitMQ med Management Plugin aktivert (`rabbitmq-plugins enable rabbitmq_management`)
- Prometheus med RabbitMQ-eksporter (valgfritt — kreves for historiske kø-grafer)

### 1. Klon repoet

```bash
git clone https://github.com/NorskHelsenett/unimq-dashboard.git
cd unimq-dashboard
```

### 2. Opprett data-mappen

```bash
mkdir -p data
```

### 3. Konfigurer tilkoblinger

Åpne filene under og oppdater tilkoblingsdetaljer for ditt miljø:

**`internal/scraper/scraper.go`**

```go
baseURL  = "http://localhost:15672/api"  // RabbitMQ Management API
username = "guest"
password = "guest"
```

**`internal/prom/prom.go`**

```go
baseURL = "http://localhost:9090/api/v1"  // Prometheus
```

**`cmd/main.go`**

```go
http.ListenAndServe(":8080", nil)  // Port dashboardet kjører på
```

### 4. Start dashboardet

```bash
go run ./cmd/main.go
```

Åpne [http://localhost:8080](http://localhost:8080) i nettleseren.

---
