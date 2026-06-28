#!/bin/sh
set -eu

CONFIG_PATH="${WATCHER_HYSTERIA_CONFIG_PATH:-/etc/hysteria/config.json}"
PID_PATH="/tmp/hysteria.pid"

start_hysteria() {
    echo "Starting Hysteria2 with config: $CONFIG_PATH"

    if [ ! -s "$CONFIG_PATH" ]; then
        echo "Hysteria config does not exist or is empty: $CONFIG_PATH"
        return 1
    fi

    echo "Testing Hysteria config..."
    hysteria check -c "$CONFIG_PATH"

    hysteria server -c "$CONFIG_PATH" &
    echo $! > "$PID_PATH"
}

stop_hysteria() {
    if [ -f "$PID_PATH" ]; then
        HYSTERIA_PID="$(cat "$PID_PATH")"
        kill "$HYSTERIA_PID" 2>/dev/null || true
        rm -f "$PID_PATH"
    fi
}

shutdown() {
    echo "Stopping..."
    stop_hysteria
    kill "$WATCHER_PID" 2>/dev/null || true
    exit 0
}

trap shutdown TERM INT

echo "Starting watcher..."
python watcher.py &
WATCHER_PID=$!

echo "Waiting for valid Hysteria config: $CONFIG_PATH"

while true; do
    if [ -s "$CONFIG_PATH" ] && hysteria check -c "$CONFIG_PATH" >/dev/null 2>&1; then
        echo "Valid Hysteria config found"
        break
    fi

    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher crashed before config was ready"
        exit 1
    fi

    sleep 2
done

start_hysteria

while true; do
    if [ -f "$PID_PATH" ]; then
        HYSTERIA_PID="$(cat "$PID_PATH")"
        if ! kill -0 "$HYSTERIA_PID" 2>/dev/null; then
            echo "Hysteria crashed or was restarted by watcher, starting again..."
            start_hysteria || true
            sleep 1
        fi
    else
        echo "Hysteria pid file missing, starting Hysteria..."
        start_hysteria || true
        sleep 1
    fi

    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher crashed, exiting..."
        stop_hysteria
        exit 1
    fi

    sleep 5
done