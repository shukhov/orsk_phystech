#!/bin/sh
set -eu

CONFIG_PATH="${XRAY_CONFIG_PATH:-/etc/xray/config.json}"
PID_PATH="/tmp/xray.pid"

start_xray() {
    echo "Starting XRay with config: $CONFIG_PATH"

    if [ ! -s "$CONFIG_PATH" ]; then
        echo "XRay config does not exist or is empty: $CONFIG_PATH"
        return 1
    fi

    echo "Testing XRay config..."
    xray run -test -c "$CONFIG_PATH"

    xray run -c "$CONFIG_PATH" &
    echo $! > "$PID_PATH"
}

stop_xray() {
    if [ -f "$PID_PATH" ]; then
        XRAY_PID="$(cat "$PID_PATH")"
        kill "$XRAY_PID" 2>/dev/null || true
        rm -f "$PID_PATH"
    fi
}

shutdown() {
    echo "Stopping..."
    stop_xray
    kill "$WATCHER_PID" 2>/dev/null || true
    exit 0
}

trap shutdown TERM INT

echo "Starting watcher..."
python watcher.py &
WATCHER_PID=$!

echo "Waiting for valid XRay config: $CONFIG_PATH"

while true; do
    if [ -s "$CONFIG_PATH" ] && xray run -test -c "$CONFIG_PATH" >/dev/null 2>&1; then
        echo "Valid XRay config found"
        break
    fi

    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher crashed before config was ready"
        exit 1
    fi

    sleep 2
done

start_xray

while true; do
    if [ -f "$PID_PATH" ]; then
        XRAY_PID="$(cat "$PID_PATH")"
        if ! kill -0 "$XRAY_PID" 2>/dev/null; then
            echo "XRay crashed or was restarted by watcher, starting again..."
            start_xray || true
            sleep 1
        fi
    else
        echo "XRay pid file missing, starting XRay..."
        start_xray || true
        sleep 1
    fi

    if ! kill -0 "$WATCHER_PID" 2>/dev/null; then
        echo "Watcher crashed, exiting..."
        stop_xray
        exit 1
    fi

    sleep 5
done