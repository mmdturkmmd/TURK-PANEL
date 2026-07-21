#!/bin/bash

set -e

PORT=${PORT:-8080}

echo "🚀 Starting Spider Panel on port ${PORT}"

mkdir -p /app/bin
mkdir -p /etc/x-ui

export XRAY_LOCATION_ASSET=/app/bin

# Migration: force webPort and subPort to match Railway PORT
if [ -f /etc/x-ui/x-ui.db ]; then
    sqlite3 /etc/x-ui/x-ui.db \
    "UPDATE settings SET value='${PORT}' WHERE key='webPort';" \
    "UPDATE settings SET value='${PORT}' WHERE key='subPort';" || true
fi

echo "✅ Xray path: ${XRAY_LOCATION_ASSET}"
echo "✅ Web port: ${PORT}"

cd /app

exec ./x-ui
