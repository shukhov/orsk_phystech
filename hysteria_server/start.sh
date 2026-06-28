#!/bin/sh
set -eu

CONFIG_PATH="${WATCHER_HYSTERIA_CONFIG_PATH:-/etc/hysteria/config.json}"
PID_PATH="/tmp/hysteria.pid"

start_hysteria() {
    echo "Starting Hysteria2 with config: $CONFIG_PATH"

    if [ ! -s "$CONFIG_PATH" ]; then
        echo "Config does not exist or is empty: $CONFIG_PATH"
        return 1
    fi

    hysteria server -c "$CONFIG_PATH" &
    echo $! > "$PID_PATH"

    # Даём процессу немного времени на запуск.
    sleep 2

    HYSTERIA_PID="$(cat "$PID_PATH")"

    if ! kill -0 "$HYSTERIA_PID" 2>/dev/null; then
        echo "Hysteria failed to start"
        rm -f "$PID_PATH"
        return 1
    fi

    echo "Hysteria started successfully (pid=$HYSTERIA_PID)"
}

stop_hysteria() {
    if [ -f "$PID_PATH" ]; then
        HYSTERIA_PID="$(cat "$PID_PATH")"
        kill "$HYSTERIA_PID" 2>/dev/null || true
        wait "$HYSTERIA_PID" 2>/dev/null || true
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

echo "Starting watcher..."
python watcher.py &
WATCHER_PID=$!

echo "Waiting for watcher to download config..."

while true; do
    if [ -s "$CONFIG_PATH" ]; then
        echo "Configuration file detected"
        break
    fi

    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher crashed before config was downloaded"
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
        echo "PID file missing, starting Hysteria..."
        start_hysteria || true
        sleep 2
        continue
    fi

    HYSTERIA_PID="$(cat "$PID_PATH")"

    if ! kill -0 "$HYSTERIA_PID" 2>/dev/null; then
        echo "Hysteria exited, restarting..."
        rm -f "$PID_PATH"
        start_hysteria || true
    fi

    sleep 5
done