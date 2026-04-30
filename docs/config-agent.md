# Configurazione Agent — Riferimento Completo

Il file di configurazione dell'agent si trova in `/etc/logsway/agent.yaml`.

Dopo ogni modifica, riavvia il servizio:
```bash
systemctl restart logsway-agent
```

---

## File di esempio completo

```yaml
# Logsway Agent Configuration

server:
  url: "http://192.168.1.10:8080"   # URL del tuo server Logsway
  timeout: 10                         # Timeout HTTP in secondi

agent:
  hostname: "web-server-01"   # Nome mostrato nella dashboard
  interval: 30                # Ogni quanti secondi inviare le metriche
  tags:
    - production
    - web

collect:
  cpu: true
  memory: true
  disk: true
  network: true
  load: true
```

---

## Riferimento campi

### `server`

| Campo | Default | Descrizione |
|-------|---------|-------------|
| `url` | — | **Obbligatorio.** URL del server Logsway inclusa la porta |
| `timeout` | `10` | Secondi prima che la richiesta HTTP venga abbandonata |

### `agent`

| Campo | Default | Descrizione |
|-------|---------|-------------|
| `hostname` | nome macchina | Nome che appare nella dashboard. Utile per nomi leggibili tipo "web-server-01" invece di "srv-prod-042.dc1.example.com" |
| `interval` | `30` | Secondi tra un invio e l'altro. Minimo consigliato: 10 |
| `tags` | `[]` | Etichette libere. Usate per raggruppare nella pagina Hosts |

**Esempi di tag:**

```yaml
# Per ambienti
tags:
  - production

# Per ruoli
tags:
  - web
  - nginx

# Misto
tags:
  - staging
  - database
  - postgres
```

### `collect`

Abilita/disabilita singole metriche. Tutti i valori sono `true` di default.

| Campo | Cosa misura |
|-------|-------------|
| `cpu` | Percentuale di utilizzo CPU |
| `memory` | Percentuale di RAM usata |
| `disk` | Percentuale di disco usata (tutte le partizioni montate) |
| `network` | Traffico di rete in/out (bytes/s) |
| `load` | Load average 1m/5m/15m |

> Se una metrica non ti serve, mettila a `false` per ridurre il traffico di rete.

---

## Esempi di configurazione per caso d'uso

### Server web in produzione

```yaml
server:
  url: "http://monitoring.interna:8080"
agent:
  hostname: "nginx-prod-01"
  interval: 15   # monitoraggio più frequente
  tags:
    - production
    - nginx
```

### Database pesante

```yaml
agent:
  hostname: "postgres-main"
  interval: 30
  tags:
    - database
    - postgres
collect:
  cpu: true
  memory: true
  disk: true
  network: false   # non mi interessa il traffico di rete
  load: true
```

### Macchina con poco spazio

```yaml
collect:
  cpu: false
  memory: false
  disk: true    # voglio SOLO monitorare il disco
  network: false
  load: false
```
