# Risoluzione Problemi Comuni

---

## L'agent non appare nella dashboard

**Sintomi:** Hai installato l'agent ma la macchina non compare nella dashboard dopo 1 minuto.

**Causa più comune:** l'agent non riesce a raggiungere il server.

**Diagnosi:**

```bash
# Sull'agent, testa la connessione al server:
curl http://IP-SERVER:8080/api/v1/health

# Risposta attesa:
{"status":"ok"}
```

Se ottieni `Connection refused` o nessuna risposta:

1. Verifica che il server Logsway sia attivo:
   ```bash
   # Sul SERVER:
   systemctl status logsway
   ```

2. Verifica che la porta sia raggiungibile:
   ```bash
   # Sul SERVER, controlla il firewall:
   ufw status
   # Se la porta 8080 non è nella lista, aprila:
   ufw allow 8080/tcp
   ```

3. Verifica l'URL nell'agent config:
   ```bash
   grep url /etc/logsway/agent.yaml
   ```

---

## Connection refused

**Il comando `curl http://SERVER:8080/api/v1/health` risponde `Connection refused`.**

Il servizio server non è in ascolto. Diagnosi:

```bash
# Sul server:
systemctl status logsway
journalctl -u logsway -n 30
```

Se il servizio risulta `failed`:

```bash
# Guarda l'errore specifico:
journalctl -u logsway --no-pager | tail -30
```

Errori comuni e soluzioni:

| Errore nel log | Soluzione |
|----------------|-----------|
| `bind: address already in use` | Un altro processo usa la porta 8080. Cambia `port` in `/etc/logsway/server.yaml` |
| `permission denied: /var/lib/logsway` | Crea la cartella: `mkdir -p /var/lib/logsway` |
| `no such file or directory: /etc/logsway/server.yaml` | Reinstalla o crea il file di config |

---

## Dashboard vuota — nessun host

**Hai agent attivi ma la dashboard mostra "No hosts yet".**

```bash
# Verifica che l'agent stia inviando:
tail -f /var/log/logsway/agent.log

# Risposta attesa (ogni ~30 secondi):
# [ok] metrics sent — cpu=5.2% ram=62.1% disk=44.0%
```

Se vedi invece `error sending metrics`:

```bash
# Testa a mano:
curl -X POST http://SERVER:8080/api/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{"hostname":"test","cpu":10,"memory":20,"disk":30}'
```

---

## Permission denied

**Errore: `permission denied` nei log del server o dell'agent.**

```bash
# Server non riesce a scrivere nel database:
ls -la /var/lib/logsway/
chown -R root:root /var/lib/logsway/
chmod 755 /var/lib/logsway/

# Agent non riesce a scrivere i log:
ls -la /var/log/logsway/
chown -R root:root /var/log/logsway/
```

---

## Dati storici non compaiono nei grafici

I grafici mostrano "No data" anche se la macchina è monitorata da ore.

**Causa:** il selettore temporale è impostato su 1h ma i dati sono stati raccolti prima.

**Soluzione:** prova a selezionare "24h" o "72h" nel grafico host detail.

---

## Reset completo — ricominciare da zero

Se qualcosa è andato storto e vuoi reinstallare pulito:

```bash
# Ferma tutto
systemctl stop logsway
systemctl stop logsway-agent 2>/dev/null || true

# Rimuovi i dati (attenzione: perdi tutto lo storico)
rm -f /var/lib/logsway/logsway.db

# Riavvia
systemctl start logsway
systemctl start logsway-agent 2>/dev/null || true
```

---

## Vedere i log in tempo reale

```bash
# Server:
tail -f /var/log/logsway/server.log

# Agent:
tail -f /var/log/logsway/agent.log

# Oppure via journalctl:
journalctl -u logsway -f
journalctl -u logsway-agent -f
```

---

## Problema non in lista?

Apri un'issue su GitHub:
https://github.com/SgtNesk/logsway-monitor/issues

Allega sempre:
```bash
systemctl status logsway
journalctl -u logsway -n 50 --no-pager
cat /etc/logsway/server.yaml
```
