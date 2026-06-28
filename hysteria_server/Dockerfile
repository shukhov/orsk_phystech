FROM python:3.12-alpine

WORKDIR /app

ARG HYSTERIA_VERSION=v2.6.2

# Устанавливаем Hysteria2 и утилиты
RUN apk add --no-cache curl ca-certificates procps && \
    curl -sL https://github.com/apernet/hysteria/releases/download/${HYSTERIA_VERSION}/hysteria-linux-amd64 -o /usr/local/bin/hysteria && \
    chmod +x /usr/local/bin/hysteria && \
    apk del curl

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY watcher.py .
COPY start.sh .
RUN chmod +x start.sh

RUN mkdir -p /etc/hysteria

CMD ["./start.sh"]
