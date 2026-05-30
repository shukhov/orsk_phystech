#!/bin/sh

sudo ufw allow 543/tcp

# Создаём пустой конфиг, если его нет
if [ ! -f /etc/xray/config.json ]; then
    echo '{}' > /etc/xray/config.json
fi

# Функция запуска XRay
start_xray() {
    echo "Starting XRay..."
    xray run -c /etc/xray/config.json &
    echo $! > /tmp/xray.pid
}

start_xray
sleep 1

echo "Starting watcher..."
python watcher.py &
WATCHER_PID=$!

# Следим за XRay — если упал, перезапускаем
while true; do
    if [ -f /tmp/xray.pid ]; then
        XRAY_PID=$(cat /tmp/xray.pid)
        if ! kill -0 $XRAY_PID 2>/dev/null; then
            echo "XRay crashed, restarting..."
            start_xray
            sleep 1
        fi
    fi
    if ! kill -0 $WATCHER_PID 2>/dev/null; then
        echo "Watcher crashed, exiting..."
        kill $(cat /tmp/xray.pid 2>/dev/null) 2>/dev/null
        exit 1
    fi
    sleep 5
done
