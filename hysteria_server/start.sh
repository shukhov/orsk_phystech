#!/bin/sh
set -eu

CONFIG_PATH="${WATCHER_HYSTERIA_CONFIG_PATH:-/etc/hysteria/config.json}"
PID_PATH="/tmp/hysteria.pid"

CERT_DIR="/etc/hysteria"
CERT_FILE="${CERT_DIR}/server.crt"
KEY_FILE="${CERT_DIR}/server.key"
PIN_FILE="${CERT_DIR}/pin.txt"

generate_certificate() {
    if [ -f "$CERT_FILE" ] && \
       [ -f "$KEY_FILE" ] && \
       [ -f "$PIN_FILE" ]; then
        echo "TLS certificate already exists"
        return
    fi

    echo "Generating self-signed certificate..."

    mkdir -p "$CERT_DIR"
    cd "$CERT_DIR"

    CERT_OUTPUT="$(
        hysteria cert \
            --cert server.crt \
            --key server.key \
            --overwrite 2>&1
    )"

    echo "$CERT_OUTPUT"

    PIN="$(
        printf "%s" "$CERT_OUTPUT" |
        grep 'pinSHA256:' |
        head -n1 |
        awk '{print $2}'
    )"

    printf "%s" "$PIN" > "$PIN_FILE"

    chmod 600 "$KEY_FILE"
    chmod 644 "$CERT_FILE" "$PIN_FILE"

    echo "pinSHA256=$(cat "$PIN_FILE")"
}

start_hysteria() {
    echo "Starting Hysteria2..."

    if [ ! -s "$CONFIG_PATH" ]; then
        echo "Configuration not found: $CONFIG_PATH"
        return 1
    fi

    hysteria server -c "$CONFIG_PATH" &
    echo $! > "$PID_PATH"

    sleep 2

    PID="$(cat "$PID_PATH")"

    if ! kill -0 "$PID" 2>/dev/null; then
        echo "Hysteria failed to start"
        rm -f "$PID_PATH"
        return 1
    fi

    echo "Hysteria started (pid=$PID)"
}

stop_hysteria() {
    if [ -f "$PID_PATH" ]; then
        PID="$(cat "$PID_PATH")"
        kill "$PID" 2>/dev/null || true
        wait "$PID" 2>/dev/null || true
        rm -f "$PID_PATH"
    fi
}

shutdown() {
    echo "Stopping..."

    stop_hysteria

    kill "$WATCHER_PID" 2>/dev/null || true
    wait "$WATCHER_PID" 2>/dev/null || true

    exit 0
}

trap shutdown TERM INT

generate_certificate

echo "Starting watcher..."

python watcher.py &
WATCHER_PID=$!

echo "Waiting for configuration..."

while true; do
    if [ -s "$CONFIG_PATH" ]; then
        echo "Configuration received"
        break
    fi

    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher exited"
        exit 1
    fi

    sleep 2
done

start_hysteria || true

while true; do
    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher exited"
        stop_hysteria
        exit 1
    fi

    if [ ! -f "$PID_PATH" ]; then
        echo "PID file missing"
        start_hysteria || true
        sleep 2
        continue
    fi

    PID="$(cat "$PID_PATH")"

    if ! kill -0 "$PID" 2>/dev/null; then
        echo "Hysteria exited"
        rm -f "$PID_PATH"
        start_hysteria || true
    fi

    sleep 5
done