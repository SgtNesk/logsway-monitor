#!/usr/bin/env bash
# Logsway Server Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/SgtNesk/logsway-monitor/main/install.sh | sudo bash

set -euo pipefail

REPO="SgtNesk/logsway-monitor"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/logsway"
DATA_DIR="/var/lib/logsway"
LOG_DIR="/var/log/logsway"
SERVICE_NAME="logsway"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
info()    { echo -e "${GREEN}[logsway]${NC} $*"; }
warning() { echo -e "${YELLOW}[warn]${NC}    $*"; }
error()   { echo -e "${RED}[error]${NC}   $*"; exit 1; }

echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║              LOGSWAY SERVER — INSTALLAZIONE               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# ── Checks ─────────────────────────────────────────────────────────────────
[[ "$EUID" -ne 0 ]] && error "Esegui come root: sudo bash $0"
command -v curl >/dev/null 2>&1 || error "curl non trovato. Installa curl prima."

# ── Detect arch ─────────────────────────────────────────────────────────────
ARCH=$(uname -m)
case $ARCH in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  *) error "Architettura non supportata: $ARCH" ;;
esac
BINARY="logsway-server-linux-${ARCH}"

# ── Get latest version ───────────────────────────────────────────────────────
info "[1/6] Cerco ultima versione..."
VERSION=$(curl -sfL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
[[ -z "$VERSION" ]] && { warning "GitHub API non raggiungibile, uso v1.0.0"; VERSION="v1.0.0"; }
info "       Trovata versione: $VERSION"

# ── Download ─────────────────────────────────────────────────────────────────
info "[2/6] Scarico ${BINARY}..."
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY}"
curl -fsSL "$DOWNLOAD_URL" -o "/tmp/${BINARY}" || error "Download fallito. URL: $DOWNLOAD_URL"
chmod +x "/tmp/${BINARY}"

# ── Install binary ────────────────────────────────────────────────────────────
info "[3/6] Installo in ${INSTALL_DIR}..."
mv "/tmp/${BINARY}" "${INSTALL_DIR}/logsway-server"

# ── Create directories ─────────────────────────────────────────────────────────
info "[4/6] Creo directory..."
mkdir -p "$CONFIG_DIR" "$DATA_DIR" "$LOG_DIR"

# ── Create config ──────────────────────────────────────────────────────────────
if [[ ! -f "${CONFIG_DIR}/server.yaml" ]]; then
  info "[5/6] Creo configurazione..."
  cat > "${CONFIG_DIR}/server.yaml" <<'EOF'
# Logsway Server Configuration

server:
  host: "0.0.0.0"
  port: 8080

database:
  path: "/var/lib/logsway/logsway.db"

retention:
  # Giorni di storico da conservare (i dati più vecchi vengono eliminati automaticamente)
  days: 30

# Soglie alert (percentuale)
thresholds:
  cpu_warning: 80
  cpu_critical: 95
  memory_warning: 85
  memory_critical: 95
  disk_warning: 80
  disk_critical: 90
EOF
else
  info "[5/6] Config esistente, mantengo..."
fi

# ── Systemd service ────────────────────────────────────────────────────────────
info "[6/6] Creo servizio systemd..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=Logsway Monitoring Server
Documentation=https://github.com/${REPO}
After=network.target

[Service]
Type=simple
User=root
ExecStart=${INSTALL_DIR}/logsway-server -config ${CONFIG_DIR}/server.yaml
Restart=always
RestartSec=5
StandardOutput=append:${LOG_DIR}/server.log
StandardError=append:${LOG_DIR}/server.log
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=${DATA_DIR} ${LOG_DIR}

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable "$SERVICE_NAME" >/dev/null 2>&1
systemctl start "$SERVICE_NAME"

IP=$(hostname -I | awk '{print $1}')
PORT=$(grep '^ *port:' "${CONFIG_DIR}/server.yaml" | awk '{print $2}' || echo 8080)

echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              INSTALLAZIONE COMPLETATA                     ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "  Dashboard:  http://${IP}:${PORT}"
echo "  Config:     ${CONFIG_DIR}/server.yaml"
echo "  Database:   ${DATA_DIR}/logsway.db"
echo "  Log:        ${LOG_DIR}/server.log"
echo "  Servizio:   systemctl status ${SERVICE_NAME}"
echo ""
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║  PROSSIMO PASSO: installa l'agent sulle macchine da       ║"
echo "║  monitorare:                                              ║"
echo "╠═══════════════════════════════════════════════════════════╣"
echo ""
echo "  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install-agent.sh | sudo bash -s -- http://${IP}:${PORT}"
echo ""
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

