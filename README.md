# LOGSWAY

Monitoring che capisci in 5 minuti, non in 5 giorni.

Logsway e un tool di infrastructure monitoring self-hosted, leggero, con agent Go, server Go + SQLite e dashboard React.

## Stack

- `agent/` — binario Go che raccoglie CPU, RAM, disk, load e network
- `server/` — API REST Go con storage SQLite embedded
- `frontend/` — dashboard React + Tailwind in stile light/minimal
- `docker-compose.yml` — avvio rapido in locale o su VPS

## Quick Start

### 1. Avvia server e dashboard

```bash
docker compose up -d --build
```

Dashboard: `http://localhost:3000`

API health: `http://localhost:8080/api/v1/health`

### 2. Avvia un agent in locale

```bash
cd agent
cp config.yaml.example config.yaml
go run . -config config.yaml
```

### 3. Apri la UI

- `/` — dashboard generale
- `/nongreen` — solo warning / critical / offline
- `/hosts` — lista host
- `/settings` — configurazione rapida lato UI

## Sviluppo locale

### Server

```bash
cd server
go mod tidy
go run .
```

Variabili utili:

- `LOGSWAY_ADDR` — bind address, default `:8080`
- `LOGSWAY_DB` — path SQLite, default `logsway.db`
- `LOGSWAY_CPU_WARN`, `LOGSWAY_CPU_CRIT`
- `LOGSWAY_RAM_WARN`, `LOGSWAY_RAM_CRIT`
- `LOGSWAY_DISK_WARN`, `LOGSWAY_DISK_CRIT`

### Agent

```bash
cd agent
cp config.yaml.example config.yaml
go mod tidy
go run . -config config.yaml
```

Config esempio:

```yaml
server:
  url: "http://localhost:8080"
  timeout: 10

agent:
  hostname: "my-server"
  interval: 30
  tags:
    - production
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## Installer

### Server

```bash
./install.sh
```

### Agent

```bash
./install-agent.sh --server http://YOUR_SERVER:8080
```

Lo script agent crea anche il servizio `systemd` quando disponibile.

## Uninstall

### Server stack

```bash
./uninstall.sh
```

### Agent

```bash
./uninstall-agent.sh
```

## Stato MVP

- Dashboard light mode con top navigation
- Grid view stile Xymon per colpo d'occhio rapido
- Pagina `/nongreen`
- Storico metriche host con grafici
- Soglie configurabili lato server via env
- SQLite embedded, nessun Postgres richiesto
