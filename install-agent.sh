#!/usr/bin/env bash
# Logsway Agent Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/SgtNesk/logsway-monitor/main/install-agent.sh | sudo bash -s -- http://SERVER:8080

set -euo pipefail

REPO="SgtNesk/logsway-monitor"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/logsway"
LOG_DIR="/var/log/logsway"
CHECKS_DIR="/etc/logsway/checks"
SERVICE_NAME="logsway-agent"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()    { echo -e "${GREEN}[agent]${NC} $*"; }
warning() { echo -e "${YELLOW}[warn]${NC}  $*"; }
error()   { echo -e "${RED}[error]${NC} $*"; exit 1; }

SERVER_URL="${1:-}"

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              LOGSWAY AGENT — INSTALLAZIONE                ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# ── Checks ─────────────────────────────────────────────────────────────────
[[ "$EUID" -ne 0 ]] && error "Esegui come root: sudo bash $0"
command -v curl >/dev/null 2>&1 || error "curl non trovato. Installa curl prima."

# ── Chiedi URL server se non passato ────────────────────────────────────────
if [[ -z "$SERVER_URL" ]]; then
  echo -e "${YELLOW}Inserisci l'URL del server Logsway${NC}"
  echo    "Esempio: http://192.168.1.10:8080"
  echo
  read -rp "URL Server: " SERVER_URL
fi
[[ -z "$SERVER_URL" ]] && error "URL server richiesto. Uso: $0 http://server:8080"

# ── Detect arch ──────────────────────────────────────────────────────────────
ARCH=$(uname -m)
case $ARCH in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) error "Architettura non supportata: $ARCH" ;;
esac
BINARY="logsway-agent-linux-${ARCH}"

# ── Get version ───────────────────────────────────────────────────────────────
info "[1/5] Cerco ultima versione..."
VERSION=$(curl -sfL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
[[ -z "$VERSION" ]] && { warning "GitHub API non raggiungibile, uso v1.0.0"; VERSION="v1.0.0"; }
info "       Versione: $VERSION"

# ── Download ──────────────────────────────────────────────────────────────────
info "[2/5] Scarico ${BINARY}..."
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"
curl -fsSL "$DOWNLOAD_URL" -o "/tmp/${BINARY}" || error "Download fallito. URL: $DOWNLOAD_URL"
chmod +x "/tmp/${BINARY}"

# ── Install ───────────────────────────────────────────────────────────────────
info "[3/5] Installo in ${INSTALL_DIR}..."
mv "/tmp/${BINARY}" "${INSTALL_DIR}/logsway-agent"

mkdir -p "$CONFIG_DIR" "$LOG_DIR" "$CHECKS_DIR"

# ── Config ────────────────────────────────────────────────────────────────────
info "[4/5] Creo configurazione..."
ACTUAL_HOSTNAME=$(hostname -f)
cat > "${CONFIG_DIR}/agent.yaml" <<EOF
# Logsway Agent Configuration

server:
  # URL del server Logsway (dove invio le metriche)
  url: "${SERVER_URL}"
  timeout: 10

# Cartella script custom checks (*.sh eseguibili)
checks_dir: "/etc/logsway/checks"

agent:
  # Nome di questa macchina (come appare nella dashboard)
  hostname: "${ACTUAL_HOSTNAME}"
  # Ogni quanti secondi invio le metriche
  interval: 30
  tags:
    - production

# Cosa monitorare (true = attivo)
collect:
  cpu: true
  memory: true
  disk: true
  network: true
  load: true
EOF

# ── Systemd service ───────────────────────────────────────────────────────────
info "[5/5] Creo servizio systemd..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=Logsway Monitoring Agent
Documentation=https://github.com/${REPO}
After=network.target

[Service]
Type=simple
User=root
ExecStart=${INSTALL_DIR}/logsway-agent -config ${CONFIG_DIR}/agent.yaml
Restart=always
RestartSec=10
StandardOutput=append:${LOG_DIR}/agent.log
StandardError=append:${LOG_DIR}/agent.log

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable "$SERVICE_NAME" >/dev/null 2>&1
systemctl start "$SERVICE_NAME"

sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
  STATUS="${GREEN}ATTIVO ✓${NC}"
else
  STATUS="${RED}ERRORE — controlla: journalctl -u ${SERVICE_NAME} -n 20${NC}"
fi

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              INSTALLAZIONE COMPLETATA                     ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "  Hostname:   ${ACTUAL_HOSTNAME}"
echo "  Server:     ${SERVER_URL}"
echo -e "  Stato:      ${STATUS}"
echo "  Config:     ${CONFIG_DIR}/agent.yaml"
echo "  Log:        ${LOG_DIR}/agent.log"
echo "  Servizio:   systemctl status ${SERVICE_NAME}"
echo ""
echo "  Questa macchina dovrebbe apparire nella dashboard"
echo "  entro 30 secondi: ${SERVER_URL}"
echo ""

