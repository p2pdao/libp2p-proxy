#!/bin/bash
set -e

# file descriptor limit for TCP connections; adjust according for your needs
ulimit -n 65536

APP_PATH="/opt/libp2p-proxy"

export GOLOG_LOG_LEVEL="info"
export GOLOG_LOG_FMT="json"
export GOLOG_FILE="$APP_PATH/libp2p-proxy.log"

exec "$APP_PATH"/libp2p-proxy -config "$APP_PATH"/server.json
