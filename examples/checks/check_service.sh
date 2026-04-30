#!/usr/bin/env bash
# check_service.sh - Controlla stato di un servizio systemd
# Uso: ./check_service.sh [service]

set -euo pipefail

SERVICE="${1:-nginx}"

if ! command -v systemctl >/dev/null 2>&1; then
  echo "critical"
  echo "0"
  echo "systemctl not available"
  exit 0
fi

if systemctl is-active --quiet "$SERVICE"; then
  echo "ok"
  echo "1"
  echo "Service $SERVICE is active"
else
  STATUS=$(systemctl is-active "$SERVICE" 2>/dev/null || echo "unknown")
  echo "critical"
  echo "0"
  echo "Service $SERVICE is not active"
  echo "Current state: $STATUS"
fi
