#!/usr/bin/env python3

import json
import logging
import os
import subprocess
import sys
import time
from datetime import datetime, timezone
from typing import Any, Optional

import requests


# ----------------------------
# Environment
# ----------------------------

API_URL = os.getenv("WATCHER_API_URL", "").rstrip("/")

WATCHER_EMAIL = os.getenv("WATCHER_EMAIL", "")
WATCHER_PASSWORD = os.getenv("WATCHER_PASSWORD", "")

POLL_INTERVAL = int(os.getenv("WATCHER_POLL_INTERVAL", "10"))

HYSTERIA_CONFIG_PATH = os.getenv("WATCHER_HYSTERIA_CONFIG_PATH", "/etc/hysteria/config.json")

# Пример:
# RESTART_COMMAND="sh -c 'test -f /tmp/hysteria.pid && kill $(cat /tmp/hysteria.pid) || true'"
RESTART_COMMAND = os.getenv("RESTART_COMMAND", "")

REQUEST_TIMEOUT = int(os.getenv("REQUEST_TIMEOUT", "15"))

# Эндпоинты
LOGIN_ENDPOINT = os.getenv("LOGIN_ENDPOINT", "/api/v1/security/login")
LAST_UPDATE_ENDPOINT = os.getenv("LAST_UPDATE_ENDPOINT", "/api/v1/hysteria/get_last_update")
CONFIG_ENDPOINT = os.getenv("CONFIG_ENDPOINT", "/api/v1/hysteria/config")

# Если API отдаёт токен не в поле "token", можно переопределить.
TOKEN_FIELD = os.getenv("TOKEN_FIELD", "token")

# Если хочешь, чтобы watcher на первом запуске всегда скачивал конфиг.
FETCH_CONFIG_ON_START = os.getenv("FETCH_CONFIG_ON_START", "true").lower() in (
    "1",
    "true",
    "yes",
    "y",
)


# ----------------------------
# Logging
# ----------------------------

logging.basicConfig(
    level=os.getenv("LOG_LEVEL", "INFO").upper(),
    format="%(asctime)s [%(levelname)s] %(message)s",
    stream=sys.stdout,
)

log = logging.getLogger("hysteria-config-watcher")


class WatcherError(Exception):
    pass


# ----------------------------
# Helpers
# ----------------------------

def require_env() -> None:
    missing = []

    if not API_URL:
        missing.append("WATCHER_API_URL")
    if not WATCHER_EMAIL:
        missing.append("WATCHER_HYSTERIA_EMAIL")
    if not WATCHER_PASSWORD:
        missing.append("WATCHER_HYSTERIA_PASSWORD")

    if missing:
        raise WatcherError(f"Missing required env vars: {', '.join(missing)}")


def url(path: str) -> str:
    if path.startswith("http://") or path.startswith("https://"):
        return path
    if not path.startswith("/"):
        path = "/" + path
    return API_URL + path


def parse_datetime(value: Any) -> Optional[datetime]:
    """
    Поддерживает:
    - ISO string: 2026-05-31T12:34:56Z
    - ISO string: 2026-05-31T12:34:56+00:00
    - unix timestamp seconds/int/float
    - null
    """
    if value is None:
        return None

    if isinstance(value, (int, float)):
        return datetime.fromtimestamp(value, tz=timezone.utc)

    if not isinstance(value, str):
        raise WatcherError(f"Unsupported datetime value type: {type(value)!r}")

    value = value.strip()
    if not value:
        return None

    # Python fromisoformat не понимает 'Z' в старых версиях.
    if value.endswith("Z"):
        value = value[:-1] + "+00:00"

    dt = datetime.fromisoformat(value)

    if dt.tzinfo is None:
        dt = dt.replace(tzinfo=timezone.utc)

    return dt.astimezone(timezone.utc)


def run_command(command: list[str], timeout: int = 30) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        command,
        capture_output=True,
        text=True,
        timeout=timeout,
    )


# ----------------------------
# API
# ----------------------------

def login() -> str:
    payload = {
        "email": WATCHER_EMAIL,
        "password": WATCHER_PASSWORD,
    }

    log.info("Logging in to API: %s", url(LOGIN_ENDPOINT))

    response = requests.post(
        url(LOGIN_ENDPOINT),
        json=payload,
        timeout=REQUEST_TIMEOUT,
    )

    if response.status_code >= 400:
        raise WatcherError(
            f"Login failed: status={response.status_code}, body={response.text}"
        )

    data = response.json()

    token = data.get(TOKEN_FIELD)
    if not token:
        raise WatcherError(
            f"Login response does not contain token field '{TOKEN_FIELD}': {data}"
        )

    log.info("Login successful")
    return token


def api_get(token: str, endpoint: str) -> Any:
    response = requests.get(
        url(endpoint),
        headers={
            "Authorization": f"Bearer {token}",
        },
        timeout=REQUEST_TIMEOUT,
    )

    if response.status_code in (401, 403):
        raise PermissionError("Token expired or unauthorized")

    if response.status_code >= 400:
        raise WatcherError(
            f"API GET failed: endpoint={endpoint}, "
            f"status={response.status_code}, body={response.text}"
        )

    return response.json()


def get_last_update(token: str) -> Optional[datetime]:
    data = api_get(token, LAST_UPDATE_ENDPOINT)

    if isinstance(data, dict):
        raw = (
            data.get("last_update")
            or data.get("lastUpdate")
            or data.get("updated_at")
            or data.get("updatedAt")
        )
    else:
        raw = data

    dt = parse_datetime(raw)
    log.debug("API last_update=%s", dt)
    return dt


def get_config(token: str) -> dict[str, Any]:
    """
    Ожидает, что API вернёт готовый Hysteria2 config JSON.
    """
    data = api_get(token, CONFIG_ENDPOINT)

    if not isinstance(data, dict):
        raise WatcherError(f"Config endpoint returned non-object JSON: {type(data)!r}")

    return data


# ----------------------------
# Hysteria config handling
# ----------------------------


def write_config_atomically(config: dict[str, Any]) -> bool:
    """
    Атомарно записывает конфиг.

    Возвращает True, если содержимое изменилось.
    """
    config_dir = os.path.dirname(HYSTERIA_CONFIG_PATH)
    config_base = os.path.basename(HYSTERIA_CONFIG_PATH)

    if config_dir:
        os.makedirs(config_dir, exist_ok=True)

    new_content = json.dumps(
        config,
        indent=2,
        ensure_ascii=False,
        sort_keys=True,
    ) + "\n"

    old_content = None
    if os.path.exists(HYSTERIA_CONFIG_PATH):
        with open(HYSTERIA_CONFIG_PATH, "r", encoding="utf-8") as f:
            old_content = f.read()

    if old_content == new_content:
        log.info("Config content is unchanged")
        return False

    tmp_path = HYSTERIA_CONFIG_PATH + ".tmp"

    with open(tmp_path, "w", encoding="utf-8") as f:
        f.write(new_content)

    os.replace(tmp_path, HYSTERIA_CONFIG_PATH)

    log.info("Config updated: %s", HYSTERIA_CONFIG_PATH)
    return True


def restart_hysteria() -> None:
    """
    RESTART_COMMAND обычно убивает Hysteria.
    start.sh замечает, что процесс умер, и запускает его заново.
    """
    if not RESTART_COMMAND:
        log.warning("RESTART_COMMAND is empty, skipping Hysteria restart")
        return

    log.info("Restarting Hysteria with command: %s", RESTART_COMMAND)

    result = subprocess.run(
        RESTART_COMMAND,
        shell=True,
        capture_output=True,
        text=True,
        timeout=30,
    )

    if result.returncode != 0:
        raise WatcherError(
            "Restart command failed\n"
            f"exit_code={result.returncode}\n"
            f"stdout={result.stdout}\n"
            f"stderr={result.stderr}"
        )

    log.info("Restart command executed successfully")


def sync_config(token: str) -> bool:
    """
    Забирает конфиг из API, валидирует, пишет, рестартит Hysteria при изменении.

    Возвращает True, если конфиг изменился и рестарт был вызван.
    """
    config = get_config(token)
    changed = write_config_atomically(config)

    if changed:
        restart_hysteria()
    else:
        log.info("Hysteria restart is not needed")

    return changed


# ----------------------------
# Main loop
# ----------------------------

def run() -> None:
    require_env()

    log.info("Starting Hysteria Config Watcher")
    log.info("API_URL=%s", API_URL)
    log.info("POLL_INTERVAL=%s", POLL_INTERVAL)
    log.info("HYSTERIA_CONFIG_PATH=%s", HYSTERIA_CONFIG_PATH)
    log.info("LOGIN_ENDPOINT=%s", LOGIN_ENDPOINT)
    log.info("LAST_UPDATE_ENDPOINT=%s", LAST_UPDATE_ENDPOINT)
    log.info("CONFIG_ENDPOINT=%s", CONFIG_ENDPOINT)

    token: Optional[str] = None
    last_seen_update: Optional[datetime] = None

    while True:
        try:
            if token is None:
                token = login()

            api_last_update = get_last_update(token)

            should_sync = False

            if FETCH_CONFIG_ON_START and last_seen_update is None:
                log.info("Initial sync is required")
                should_sync = True

            elif api_last_update is None:
                log.warning("API returned empty last_update; skipping sync")

            elif last_seen_update is None:
                log.info("No local last_seen_update, syncing config")
                should_sync = True

            elif api_last_update > last_seen_update:
                log.info(
                    "Detected config update: previous=%s current=%s",
                    last_seen_update,
                    api_last_update,
                )
                should_sync = True

            else:
                log.debug(
                    "No changes detected: last_seen_update=%s api_last_update=%s",
                    last_seen_update,
                    api_last_update,
                )

            if should_sync:
                sync_config(token)

                # Важно: обновляем last_seen_update после успешной записи/валидации.
                last_seen_update = api_last_update or datetime.now(timezone.utc)

        except PermissionError as e:
            log.warning("Authorization problem: %s. Re-login on next iteration.", e)
            token = None

        except requests.ConnectionError as e:
            log.error("API connection error: %s", e)
            token = None

        except requests.Timeout as e:
            log.error("API timeout: %s", e)

        except WatcherError as e:
            log.error("Watcher error: %s", e)

        except Exception as e:
            log.error("Unexpected error: %s", e, exc_info=True)

        time.sleep(POLL_INTERVAL)


if __name__ == "__main__":
    try:
        run()
    except WatcherError as e:
        log.error("Fatal error: %s", e)
        sys.exit(1)