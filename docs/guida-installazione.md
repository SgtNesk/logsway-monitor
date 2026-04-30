# Guida Installazione Logsway

Questa guida ti accompagna passo-passo nell'installazione di Logsway.
Non serve esperienza precedente con il monitoring.

## Indice

1. [Cos'è Logsway](#cosè-logsway)
2. [Come funziona (architettura)](#come-funziona)
3. [Requisiti](#requisiti)
4. [Installazione Server](#installazione-server)
5. [Installazione Agent](#installazione-agent)
6. [Verifica](#verifica)
7. [Comandi utili](#comandi-utili)
8. [Custom Checks](#custom-checks)
9. [Prossimi passi](#prossimi-passi)

---

## Cos'è Logsway

Logsway monitora i tuoi server. Ti dice:
- Quanta CPU stanno usando
- Quanta RAM è occupata
- Quanto spazio disco rimane
- Se qualcosa non va (alert)

È composto da due pezzi:

| Componente | Dove si installa | Cosa fa |
|------------|------------------|---------|
| **Server** | 1 macchina centrale | Raccoglie i dati, mostra la dashboard |
| **Agent** | Ogni macchina da monitorare | Legge le metriche, le invia al server |

---

## Come funziona

```
    LE TUE MACCHINE                        DASHBOARD

    ┌─────────────────┐
    │  web-server-01  │
    │    [Agent]      │────┐
    └─────────────────┘    │
                           │     ┌─────────────────────┐
    ┌─────────────────┐    │     │                     │
    │  web-server-02  │    ├────▶│  LOGSWAY SERVER     │◀── Browser
    │    [Agent]      │    │     │                     │
    └─────────────────┘    │     └─────────────────────┘
                           │
    ┌─────────────────┐    │
    │  database-01    │────┘
    │    [Agent]      │
    └─────────────────┘
```

**Il flusso è semplice:**
1. L'**Agent** legge CPU/RAM/Disk della macchina ogni 30 secondi
2. L'**Agent** invia i dati al **Server** via HTTP
3. Il **Server** salva tutto nel database
4. Tu apri il browser → vedi tutto in tempo reale

---

## Requisiti

### Per il Server (1 sola macchina)

| Requisito | Minimo |
|-----------|--------|
| OS | Ubuntu 20.04+ / Debian 11+ |
| RAM | 512 MB |
| Disco | 1 GB |
| Rete | Porta 8080 raggiungibile dagli agent |

### Per ogni Agent

| Requisito | Minimo |
|-----------|--------|
| OS | Linux (qualsiasi distribuzione moderna) |
| RAM | 32 MB |
| Rete | Raggiunge il server sulla porta 8080 |

---

## Installazione Server

### Passo 1: Accedi alla macchina server

```bash
ssh tuo-utente@indirizzo-server
# Esempio: ssh admin@192.168.1.10
```

> **Cos'è SSH?** Un modo sicuro per collegarti a un computer remoto via terminale.

### Passo 2: Esegui l'installer

```bash
curl -fsSL https://raw.githubusercontent.com/SgtNesk/logsway-monitor/main/install.sh | sudo bash
```

> **Cosa fa questo comando:**
> - `curl` scarica lo script da internet
> - `sudo bash` lo esegue come amministratore

### Passo 3: Attendi il completamento

Vedrai 6 step e alla fine:

```
╔═══════════════════════════════════════╗
║     INSTALLAZIONE COMPLETATA          ║
╚═══════════════════════════════════════╝

  Dashboard:  http://192.168.1.10:8080
```

### Passo 4: Apri la dashboard

Vai all'indirizzo mostrato. Vedrai la dashboard vuota — nessun host ancora.

---

## Installazione Agent

Ripeti questi passi su **ogni macchina** che vuoi monitorare.

### Passo 1: Accedi alla macchina

```bash
ssh tuo-utente@macchina-da-monitorare
```

### Passo 2: Esegui l'installer

Sostituisci `http://192.168.1.10:8080` con l'indirizzo del TUO server:

```bash
curl -fsSL https://raw.githubusercontent.com/SgtNesk/logsway-monitor/main/install-agent.sh | sudo bash -s -- http://192.168.1.10:8080
```

### Passo 3: Verifica

Torna alla dashboard. **Entro 30 secondi** vedrai apparire la nuova macchina.

---

## Verifica

### Controllare che il server funzioni

```bash
systemctl status logsway
```

Cerca `Active: active (running)` — verde.

### Controllare che l'agent funzioni

```bash
systemctl status logsway-agent
tail -20 /var/log/logsway/agent.log
```

Dovresti vedere righe tipo:
```
[ok] metrics sent — cpu=12.3% ram=45.6% disk=22.1%
```

### Test connessione

Dall'agent, prova a raggiungere il server:

```bash
curl http://192.168.1.10:8080/api/v1/health
# Output atteso: {"status":"ok"}
```

---

## Comandi Utili

### Gestione servizi

| Azione | Server | Agent |
|--------|--------|-------|
| Stato | `systemctl status logsway` | `systemctl status logsway-agent` |
| Avvia | `systemctl start logsway` | `systemctl start logsway-agent` |
| Ferma | `systemctl stop logsway` | `systemctl stop logsway-agent` |
| Riavvia | `systemctl restart logsway` | `systemctl restart logsway-agent` |
| Log live | `tail -f /var/log/logsway/server.log` | `tail -f /var/log/logsway/agent.log` |

### Modificare configurazione

```bash
# Server
nano /etc/logsway/server.yaml
systemctl restart logsway

# Agent
nano /etc/logsway/agent.yaml
systemctl restart logsway-agent
```

### Disinstallare

```bash
# Server
curl -fsSL https://raw.githubusercontent.com/SgtNesk/logsway-monitor/main/uninstall.sh | sudo bash

# Agent
curl -fsSL https://raw.githubusercontent.com/SgtNesk/logsway-monitor/main/uninstall-agent.sh | sudo bash
```

---

## Custom Checks

Puoi monitorare qualsiasi cosa scrivendo script bash. L'agent esegue automaticamente ogni
script `.sh` eseguibile in `/etc/logsway/checks` e invia i risultati al server.

### Creare un check

1. Crea lo script:

```bash
sudo mkdir -p /etc/logsway/checks
sudo nano /etc/logsway/checks/check_backup.sh
```

2. Formato output (3 blocchi):

```text
ok|warning|critical
valore_numerico
messaggio dettagliato (una o piu righe)
```

3. Rendi eseguibile:

```bash
sudo chmod +x /etc/logsway/checks/check_backup.sh
```

4. Fine. Entro ~30 secondi appare in [Matrix](http://localhost:8080/matrix)
come nuova colonna con nome `backup` (da `check_backup.sh`).

### Esempio minimo

```bash
#!/usr/bin/env bash
if [[ -f /var/backups/latest.tar.gz ]]; then
  echo "ok"
  echo "1"
  echo "Backup exists"
else
  echo "critical"
  echo "0"
  echo "Backup file missing!"
fi
```

### Script pronti da copiare

Nella repo trovi esempi in `examples/checks/`:
- `check_backup.sh`
- `check_ssl.sh`
- `check_service.sh`
- `check_disk_smart.sh`

Puoi copiarli cosi:

```bash
sudo mkdir -p /etc/logsway/checks
sudo cp examples/checks/*.sh /etc/logsway/checks/
sudo chmod +x /etc/logsway/checks/*.sh
```

### Test rapido

```bash
cat <<'EOF' | sudo tee /etc/logsway/checks/check_test.sh >/dev/null
#!/usr/bin/env bash
echo "ok"
echo "42"
echo "Test check working"
EOF
sudo chmod +x /etc/logsway/checks/check_test.sh
sudo systemctl restart logsway-agent
```

Poi apri `/matrix`: deve apparire la colonna `test`.

---

## Prossimi passi

- [Configurare le soglie alert](config-server.md)
- [Risolvere problemi comuni](troubleshooting.md)
- [Installazione con Docker](docker.md)
