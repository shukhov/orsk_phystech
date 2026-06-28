FROM python:3.12-alpine

WORKDIR /app

ARG HYSTERIA_VERSION=v2.6.2

# Устанавливаем Hysteria2 и утилиты (с автоматическим определением архитектуры)
RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        x86_64)  BINARY="hysteria-linux-amd64" ;; \
        aarch64) BINARY="hysteria-linux-arm64" ;; \
        armv7l)  BINARY="hysteria-linux-arm" ;; \
        *)       echo "Unsupported architecture: $ARCH"; exit 1 ;; \
    esac && \
    echo "Downloading ${BINARY}" && \
    apk add --no-cache curl ca-certificates procps && \
    curl -sL https://github.com/apernet/hysteria/releases/download/${HYSTERIA_VERSION}/${BINARY} -o /usr/local/bin/hysteria && \
    chmod +x /usr/local/bin/hysteria && \
    apk del curl

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY watcher.py .
COPY start.sh .
RUN chmod +x start.sh

RUN mkdir -p /etc/hysteria

CMD ["./start.sh"]
