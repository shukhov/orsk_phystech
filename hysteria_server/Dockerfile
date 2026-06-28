FROM python:3.12-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        curl \
        ca-certificates \
        procps && \
    curl -fsSL \
        https://download.hysteria.network/app/latest/hysteria-linux-amd64 \
        -o /usr/local/bin/hysteria && \
    chmod +x /usr/local/bin/hysteria && \
    hysteria version && \
    apt-get purge -y curl && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY watcher.py .
COPY start.sh .
RUN chmod +x start.sh

RUN mkdir -p /etc/hysteria

CMD ["./start.sh"]