FROM python:3.12-alpine

WORKDIR /app

# Устанавливаем Hysteria2 и утилиты
RUN apk add --no-cache curl procps && \
    HYSTERIA_VERSION=$(curl -sL https://api.github.com/repos/apernet/hysteria/releases/latest | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/') && \
    curl -sL https://github.com/apernet/hysteria/releases/download/v${HYSTERIA_VERSION}/hysteria-linux-amd64 -o /usr/local/bin/hysteria && \
    chmod +x /usr/local/bin/hysteria && \
    apk del curl

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY watcher.py .
COPY start.sh .
RUN chmod +x start.sh

RUN mkdir -p /etc/hysteria

CMD ["./start.sh"]