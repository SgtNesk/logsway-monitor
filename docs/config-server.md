# Configurazione Server — Riferimento Completo

Il file di configurazione del server si trova in `/etc/logsway/server.yaml`.

Dopo ogni modifica, riavvia il servizio:
```bash
systemctl restart logsway
```

---

## File di esempio completo

```yaml
# Logsway Server Configuration

server:
  host: "0.0.0.0"   # Ascolta su tutte le interfacce (default)
  port: 8080          # Porta HTTP

database:
  path: "/var/lib/logsway/logsway.db"   # Path del database SQLite

retention:
  days: 30   # Conserva 30 giorni di storico

thresholds:
  cpu_warning: 80     # % CPU → giallo
  cpu_critical: 95    # % CPU → rosso
  memory_warning: 85
  memory_critical: 95
  disk_warning: 80
  disk_critical: 90
```

---

## Riferimento campi

### `server`

| Campo | Default | Descrizione |
|-------|---------|-------------|
| `host` | `"0.0.0.0"` | Indirizzo su cui ascolta. `0.0.0.0` = tutte le interfacce |
| `port` | `8080` | Porta HTTP. Deve essere libera sul server |

> Se cambi la porta, aggiorna anche il firewall:
> ```bash
> ufw allow 9090/tcp   # esempio con porta 9090
> ```

### `database`

| Campo | Default | Descrizione |
|-------|---------|-------------|
| `path` | `/var/lib/logsway/logsway.db` | Path del database SQLite. La cartella deve esistere |

### `retention`

| Campo | Default | Descrizione |
|-------|---------|-------------|
| `days` | `30` | Giorni di dati da conservare. I dati più vecchi vengono eliminati automaticamente |

> Con 10 agent che inviano ogni 30s per 30 giorni → circa 200 MB di database.

### `thresholds`

Le soglie determinano il colore degli host nella dashboard:

| Campo | Default | Descrizione |
|-------|---------|-------------|
| `cpu_warning` | `80` | CPU oltre questa % → giallo |
| `cpu_critical` | `95` | CPU oltre questa % → rosso |
| `memory_warning` | `85` | RAM oltre questa % → giallo |
| `memory_critical` | `95` | RAM oltre questa % → rosso |
| `disk_warning` | `80` | Disco oltre questa % → giallo |
| `disk_critical` | `90` | Disco oltre questa % → rosso |

---

## Sicurezza: esporre Logsway su internet

Se accedi alla dashboard da internet (non solo dalla LAN), metti un reverse proxy davanti con autenticazione:

```nginx
# /etc/nginx/sites-available/logsway
server {
    listen 443 ssl;
    server_name monitoring.tua-azienda.it;

    auth_basic "Logsway";
    auth_basic_user_file /etc/nginx/.logsway-passwd;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Crea la password:
```bash
htpasswd -c /etc/nginx/.logsway-passwd admin
```
