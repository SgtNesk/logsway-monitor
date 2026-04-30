#!/usr/bin/env bash
# Logsway Alert Script
# Controlla gli host con problemi e invia notifiche
#
# Uso: ./alert.sh [SERVER_URL]
# Esempio: ./alert.sh http://192.168.1.10:8080
#
# Integrazione cron (ogni 5 minuti):
#   */5 * * * * /usr/local/bin/logsway-alert.sh http://localhost:8080

set -euo pipefail

# ─── Configurazione ──────────────────────────────────────────────────────────

SERVER_URL="${1:-http://localhost:8080}"

# Email (lascia vuoto per disabilitare)
ALERT_EMAIL=""
# Esempio: ALERT_EMAIL="ops-team@tuaazienda.it"

# Slack webhook (lascia vuoto per disabilitare)
SLACK_WEBHOOK=""
# Esempio: SLACK_WEBHOOK="https://hooks.slack.com/services/T.../B.../..."

# Telegram (lascia entrambi vuoti per disabilitare)
TELEGRAM_BOT_TOKEN=""
TELEGRAM_CHAT_ID=""

# File di log
LOG_FILE="/var/log/logsway/alerts.log"
# File di stato per evitare notifiche duplicate
STATE_DIR="/var/lib/logsway/alert-state"

# ─── Setup ───────────────────────────────────────────────────────────────────

mkdir -p "$STATE_DIR"

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "$LOG_FILE" 2>/dev/null || true
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# ─── Funzioni di invio ────────────────────────────────────────────────────────

send_email() {
  local subject="$1"
  local body="$2"
  [[ -z "$ALERT_EMAIL" ]] && return 0
  command -v mail >/dev/null 2>&1 || { log "[warn] mail non installato, skip email"; return 0; }
  echo "$body" | mail -s "$subject" "$ALERT_EMAIL" && log "[email] Inviata a $ALERT_EMAIL"
}

send_slack() {
  local msg="$1"
  [[ -z "$SLACK_WEBHOOK" ]] && return 0
  command -v curl >/dev/null 2>&1 || return 0
  curl -sf -X POST "$SLACK_WEBHOOK" \
    -H 'Content-type: application/json' \
    -d "{\"text\":\"${msg}\"}" >/dev/null && log "[slack] Inviato"
}

send_telegram() {
  local msg="$1"
  [[ -z "$TELEGRAM_BOT_TOKEN" || -z "$TELEGRAM_CHAT_ID" ]] && return 0
  command -v curl >/dev/null 2>&1 || return 0
  curl -sf -X POST "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/sendMessage" \
    -d "chat_id=${TELEGRAM_CHAT_ID}" \
    -d "text=${msg}" \
    -d "parse_mode=HTML" >/dev/null && log "[telegram] Inviato"
}

notify() {
  local host="$1"
  local severity="$2"   # warning | critical | offline
  local metric="$3"
  local value="$4"
  local threshold="$5"

  # Evita notifiche duplicate: controlla lo stato precedente
  STATE_FILE="${STATE_DIR}/${host}.state"
  PREV_STATE=$(cat "$STATE_FILE" 2>/dev/null || echo "ok")

  if [[ "$PREV_STATE" == "$severity" ]]; then
    # Stesso stato di prima → non inviare di nuovo
    return 0
  fi

  echo "$severity" > "$STATE_FILE"

  local EMOJI="⚠️"
  [[ "$severity" == "critical" || "$severity" == "offline" ]] && EMOJI="🔴"

  local subject="${EMOJI} [Logsway] ${severity^^}: ${host}"
  local body
  body="Host: ${host}
Severity: ${severity}
Problema: ${metric} = ${value} (soglia: ${threshold})
Timestamp: $(date)
Dashboard: ${SERVER_URL}/hosts/${host}"

  log "[alert] ${severity^^}: ${host} — ${metric}=${value}"
  send_email "$subject" "$body"
  send_slack "${EMOJI} *${severity^^}* — \`${host}\`\n${metric} = ${value} (soglia: ${threshold})\n${SERVER_URL}/hosts/${host}"
  send_telegram "${EMOJI} <b>${severity^^}</b>: <code>${host}</code>\n${metric} = ${value}\n<a href='${SERVER_URL}/hosts/${host}'>Apri dashboard</a>"
}

# ─── Recupero host con problemi ────────────────────────────────────────────────

HOSTS_JSON=$(curl -sf "${SERVER_URL}/api/v1/hosts" 2>/dev/null) || {
  log "[error] Impossibile contattare il server: ${SERVER_URL}"
  exit 1
}

# Analizza con python3 o awk (python3 è più affidabile per il parsing JSON)
if command -v python3 >/dev/null 2>&1; then
  python3 - <<PYEOF
import json, subprocess, sys, os

hosts = json.loads('''${HOSTS_JSON}''')
state_dir = "${STATE_DIR}"

for h in hosts:
    name = h.get("hostname", "unknown")
    status = h.get("status", "ok")
    cpu = h.get("cpu", 0)
    ram = h.get("memory", 0)
    disk = h.get("disk", 0)

    if status in ["warning", "critical", "offline"]:
        # Determina metrica principale
        metric, value, threshold = "status", status, "-"
        if status != "offline":
            if cpu >= 80:
                metric, value, threshold = "CPU", f"{cpu:.1f}%", "80%"
            elif ram >= 85:
                metric, value, threshold = "RAM", f"{ram:.1f}%", "85%"
            elif disk >= 80:
                metric, value, threshold = "Disk", f"{disk:.1f}%", "80%"

        subprocess.run([
            "/bin/bash", "-c",
            f'source {os.path.abspath(__file__) if False else sys.argv[0]}; notify "{name}" "{status}" "{metric}" "{value}" "{threshold}"'
        ])
    else:
        # Host tornato ok → resetta stato
        state_file = os.path.join(state_dir, f"{name}.state")
        if os.path.exists(state_file):
            with open(state_file, "r") as f:
                prev = f.read().strip()
            if prev != "ok":
                print(f"[{__import__('datetime').datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] [ok] {name} tornato ok")
                with open(state_file, "w") as f:
                    f.write("ok")
PYEOF
else
  # Fallback senza python3: usa grep/awk per trovare host non-ok
  PROBLEMS=$(echo "$HOSTS_JSON" | grep -oP '"hostname":"[^"]+","[^}]+"status":"(warning|critical|offline)"' | grep -oP '"hostname":"[^"]+"' | cut -d'"' -f4 || true)
  if [[ -n "$PROBLEMS" ]]; then
    log "[alert] Host con problemi: $PROBLEMS (installa python3 per notifiche dettagliate)"
    notify "multiple-hosts" "warning" "vedi dashboard" "-" "-"
  fi
fi

log "[check] Completato"
