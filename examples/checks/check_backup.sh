#!/usr/bin/env bash
# check_backup.sh - Controlla età ultimo backup
# Uso: ./check_backup.sh [/var/backups/latest.tar.gz]

set -euo pipefail

BACKUP_FILE="${1:-/var/backups/latest.tar.gz}"
LAST_BACKUP=$(stat -c %Y "$BACKUP_FILE" 2>/dev/null || echo 0)
NOW=$(date +%s)
AGE_HOURS=$(( (NOW - LAST_BACKUP) / 3600 ))

if [[ "$LAST_BACKUP" -eq 0 ]]; then
  echo "critical"
  echo "9999"
  echo "Backup file not found"
  echo "File: $BACKUP_FILE"
  exit 0
fi

if [[ "$AGE_HOURS" -gt 48 ]]; then
  echo "critical"
  echo "$AGE_HOURS"
  echo "Last backup: $AGE_HOURS hours ago"
  echo "File: $BACKUP_FILE"
  echo "CRITICAL: Backup too old"
elif [[ "$AGE_HOURS" -gt 24 ]]; then
  echo "warning"
  echo "$AGE_HOURS"
  echo "Last backup: $AGE_HOURS hours ago"
  echo "WARNING: Backup getting stale"
else
  echo "ok"
  echo "$AGE_HOURS"
  echo "Last backup: $AGE_HOURS hours ago"
  echo "All good"
fi
