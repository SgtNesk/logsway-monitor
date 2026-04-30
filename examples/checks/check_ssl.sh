#!/usr/bin/env bash
# check_ssl.sh - Controlla giorni rimanenti certificato SSL
# Uso: ./check_ssl.sh [domain]

set -euo pipefail

DOMAIN="${1:-localhost}"
RAW_ENDDATE=$(echo | openssl s_client -servername "$DOMAIN" -connect "$DOMAIN":443 2>/dev/null | openssl x509 -noout -enddate 2>/dev/null | cut -d= -f2 || true)

if [[ -z "$RAW_ENDDATE" ]]; then
  echo "critical"
  echo "0"
  echo "Unable to read SSL certificate for $DOMAIN"
  exit 0
fi

EXP_TS=$(date -d "$RAW_ENDDATE" +%s 2>/dev/null || echo 0)
NOW_TS=$(date +%s)
DAYS_LEFT=$(( (EXP_TS - NOW_TS) / 86400 ))

if [[ "$DAYS_LEFT" -lt 0 ]]; then
  echo "critical"
  echo "0"
  echo "Certificate already expired for $DOMAIN"
elif [[ "$DAYS_LEFT" -lt 7 ]]; then
  echo "critical"
  echo "$DAYS_LEFT"
  echo "Certificate expires in $DAYS_LEFT days for $DOMAIN"
elif [[ "$DAYS_LEFT" -lt 21 ]]; then
  echo "warning"
  echo "$DAYS_LEFT"
  echo "Certificate expires in $DAYS_LEFT days for $DOMAIN"
else
  echo "ok"
  echo "$DAYS_LEFT"
  echo "Certificate valid for $DAYS_LEFT days for $DOMAIN"
fi
