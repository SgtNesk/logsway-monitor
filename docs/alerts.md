# Alert e Notifiche

Logsway non invia email/Slack direttamente — usa un meccanismo semplice e flessibile: quando una metrica supera una soglia, puoi configurare uno **script di alert**.

---

## Come funziona

Le soglie sono definite in `/etc/logsway/server.yaml`:

```yaml
thresholds:
  cpu_warning: 80
  cpu_critical: 95
  memory_warning: 85
  memory_critical: 95
  disk_warning: 80
  disk_critical: 90
```

Quando un host supera una soglia, cambia stato nella dashboard. Puoi fare polling dell'API e triggerare azioni custom.

---

## Script di alert pronto all'uso

Nella cartella `examples/` trovi `alert.sh` — uno script da integrare con cron o systemd timer.

```bash
# Copia lo script
cp examples/alert.sh /usr/local/bin/logsway-alert.sh
chmod +x /usr/local/bin/logsway-alert.sh

# Configura le variabili (email, Slack, etc.)
nano /usr/local/bin/logsway-alert.sh
```

### Aggiungi al crontab

```bash
crontab -e
```

```cron
# Controlla alert ogni 5 minuti
*/5 * * * * /usr/local/bin/logsway-alert.sh http://localhost:8080 >> /var/log/logsway/alert-cron.log 2>&1
```

---

## Notifiche via email

Assicurati che `mail` sia installato:

```bash
# Debian/Ubuntu
apt install mailutils -y

# Test
echo "Test" | mail -s "Test email" tuo@email.com
```

Configura il server SMTP in `/etc/ssmtp/ssmtp.conf` (o usa postfix/sendmail).

---

## Notifiche via Slack

Crea un webhook su: https://api.slack.com/apps → Incoming Webhooks

Copia il webhook URL (formato: `https://hooks.slack.com/services/...`) e impostalo in `alert.sh`:

```bash
SLACK_WEBHOOK="https://hooks.slack.com/services/T.../B.../..."
```

---

## Notifiche via Telegram

Per Telegram, usa il Bot API:

```bash
TELEGRAM_BOT_TOKEN="12345:AAB..."
TELEGRAM_CHAT_ID="-1001234567890"

send_telegram() {
  local msg="$1"
  curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
    -d "chat_id=${TELEGRAM_CHAT_ID}" \
    -d "text=${msg}" \
    -d "parse_mode=HTML" >/dev/null
}
```

---

## API per polling personalizzato

Puoi integrare Logsway con qualsiasi sistema tramite le API:

```bash
# Lista tutti gli host con stato
curl http://localhost:8080/api/v1/hosts | jq '.[].status'

# Host in warning o critical
curl http://localhost:8080/api/v1/hosts | jq '[.[] | select(.status == "warning" or .status == "critical")]'

# Statistiche globali
curl http://localhost:8080/api/v1/stats
```

### Esempio: script bash semplice

```bash
#!/usr/bin/env bash
SERVER="http://localhost:8080"

# Ottieni host con problemi
PROBLEMATIC=$(curl -sf "${SERVER}/api/v1/hosts" | \
  python3 -c "
import json, sys
hosts = json.load(sys.stdin)
problems = [h for h in hosts if h.get('status') in ['warning', 'critical', 'offline']]
for h in problems:
    print(f\"{h['status'].upper()}: {h['hostname']} — CPU {h.get('cpu', 0):.1f}%\")
")

if [[ -n "$PROBLEMATIC" ]]; then
  echo "$PROBLEMATIC" | mail -s "[Logsway] Alert" admin@tuaazienda.it
fi
```
