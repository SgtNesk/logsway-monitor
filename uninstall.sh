#!/usr/bin/env bash
# Logsway Server Uninstaller

set -euo pipefail

GREEN='\033[0;32m'; RED='\033[0;31m'; NC='\033[0m'

echo "Stopping and removing Logsway server..."

systemctl stop logsway 2>/dev/null || true
systemctl disable logsway 2>/dev/null || true
rm -f /etc/systemd/system/logsway.service
systemctl daemon-reload || true

rm -f /usr/local/bin/logsway-server

echo -e "${GREEN}Logsway server rimosso.${NC}"
echo "  Dati conservati in: /var/lib/logsway/"
echo "  Config conservata in: /etc/logsway/"
echo "  Per rimuoverli completamente: rm -rf /var/lib/logsway /etc/logsway /var/log/logsway"
