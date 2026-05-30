FROM python:3.12-alpine

WORKDIR /app

# Устанавливаем XRay-core и утилиты
RUN apk add --no-cache curl unzip procps && \
    XRAY_VERSION=$(curl -sL https://api.github.com/repos/XTLS/Xray-core/releases/latest | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/') && \
    curl -sL https://github.com/XTLS/Xray-core/releases/download/v${XRAY_VERSION}/Xray-linux-64.zip -o xray.zip && \
    unzip xray.zip -d /usr/local/bin xray && \
    chmod +x /usr/local/bin/xray && \
    rm xray.zip && \
    apk del curl unzip

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY watcher.py .
COPY start.sh .
RUN chmod +x start.sh

RUN mkdir -p /etc/xray

CMD ["./start.sh"]
