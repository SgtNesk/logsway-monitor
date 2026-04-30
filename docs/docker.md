# Installazione con Docker

Alternativa all'installazione con binario. Utile se hai già Docker sul server.

---

## Prerequisiti

- Docker 20.10+
- Docker Compose v2

```bash
docker --version
docker compose version
```

---

## Avvio rapido

```bash
# Clona il repository
git clone https://github.com/SgtNesk/logsway-monitor.git
cd logsway-monitor

# Avvia
docker compose up -d
```

La dashboard sarà disponibile su `http://localhost:3000`.

---

## Struttura docker-compose.yml

Il file preconfigurato include:
- **server**: backend Go sulla porta 8080
- **frontend**: nginx con la UI React sulla porta 3000
- **logsway-data**: volume persistente per il database

---

## Configurazione

Per personalizzare, crea un file `.env` nella stessa cartella:

```bash
# .env
LOGSWAY_PORT=8080
LOGSWAY_UI_PORT=3000
```

Oppure monta un file di config personalizzato:

```yaml
# docker-compose.override.yml
services:
  server:
    volumes:
      - ./my-server.yaml:/etc/logsway/server.yaml:ro
```

---

## Agent con Docker

Puoi eseguire l'agent in un container, ma ha accesso limitato alle metriche host.
**Consiglio:** installa l'agent come binario nativo per avere tutte le metriche.

Se vuoi comunque usarlo con Docker:

```yaml
# docker-compose.yml parziale per l'agent
services:
  agent:
    image: logsway-agent:latest
    network_mode: host    # necessario per leggere le metriche host
    environment:
      - SERVER_URL=http://server:8080
      - HOSTNAME=my-server
```

---

## Comandi utili

```bash
# Stato
docker compose ps

# Log server
docker compose logs -f server

# Log frontend
docker compose logs -f frontend

# Ferma tutto
docker compose down

# Ferma e rimuovi i dati (attenzione!)
docker compose down -v
```

---

## Backup del database

Il database è nel volume `logsway-data`. Per fare backup:

```bash
# Trova il path del volume
docker volume inspect logsway-monitor_logsway-data

# Copia il database
docker run --rm \
  -v logsway-monitor_logsway-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/logsway-backup-$(date +%Y%m%d).tar.gz /data
```
