#!/usr/bin/env bash
# Logsway Agent Uninstaller

set -euo pipefail

GREEN='\033[0;32m'; NC='\033[0m'

echo "Stopping and removing Logsway agent..."

systemctl stop logsway-agent 2>/dev/null || true
systemctl disable logsway-agent 2>/dev/null || true
rm -f /etc/systemd/system/logsway-agent.service
systemctl daemon-reload || true

rm -f /usr/local/bin/logsway-agent

echo -e "${GREEN}Logsway agent rimosso.${NC}"
echo "  Config conservata in: /etc/logsway/agent.yaml"
echo "  Per rimuoverla: rm -f /etc/logsway/agent.yaml"
