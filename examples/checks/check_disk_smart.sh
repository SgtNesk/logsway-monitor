#!/usr/bin/env bash
# check_disk_smart.sh - Controlla salute SMART del disco
# Uso: ./check_disk_smart.sh [/dev/sda]

set -euo pipefail

DISK="${1:-/dev/sda}"

if ! command -v smartctl >/dev/null 2>&1; then
  echo "warning"
  echo "0"
  echo "smartctl not installed (install smartmontools to enable SMART checks)"
  exit 0
fi

OUT=$(smartctl -H "$DISK" 2>/dev/null || true)
if [[ -z "$OUT" ]]; then
  echo "critical"
  echo "0"
  echo "Unable to read SMART health for $DISK"
  exit 0
fi

if echo "$OUT" | grep -qi "PASSED"; then
  echo "ok"
  echo "1"
  echo "SMART health PASSED for $DISK"
else
  echo "critical"
  echo "0"
  echo "SMART health check FAILED for $DISK"
  echo "$OUT"
fi
