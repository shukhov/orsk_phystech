#!/usr/bin/env python3
"""
XRay Config Watcher.

Периодически проверяет /xray/last_update и при изменениях
обновляет конфигурацию XRay-сервера.
"""

import os
import sys
import json
import time
import logging
import subprocess
from datetime import datetime, timezone, timedelta

import requests

# --- Настройка логирования ---
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S',
)
log = logging.getLogger('watcher')

# --- Переменные окружения ---
API_URL = os.environ.get('API_URL', 'http://api:8081')
LOGIN_EMAIL = os.environ.get('LOGIN_EMAIL', '')
LOGIN_PASSWORD = os.environ.get('LOGIN_PASSWORD', '')
POLL_INTERVAL = int(os.environ.get('POLL_INTERVAL', '60'))       # секунд
THRESHOLD_MINUTES = int(os.environ.get('THRESHOLD_MINUTES', '1')) # минут
XRAY_CONFIG_PATH = os.environ.get('XRAY_CONFIG_PATH', '/etc/xray/config.json')
RESTART_COMMAND = os.environ.get('RESTART_COMMAND', '')           # команда для перезапуска xray


class APIError(Exception):
    """Ошибка при обращении к API."""
    pass


def login() -> str:
    """Авторизоваться и получить JWT-токен."""
    url = f'{API_URL}/api/v1/security/login'
    resp = requests.post(url, json={
        'email': LOGIN_EMAIL,
        'password': LOGIN_PASSWORD,
    }, timeout=10)
    if resp.status_code != 201:
        raise APIError(f'Login failed: {resp.status_code} {resp.text}')
    token = resp.json().get('token')
    if not token:
        raise APIError(f'No token in login response: {resp.text}')
    log.info('Успешная авторизация')
    return token


def api_get(path: str, token: str) -> dict:
    """GET-запрос к API с авторизацией."""
    url = f'{API_URL}{path}'
    headers = {'Authorization': f'Bearer {token}'}
    resp = requests.get(url, headers=headers, timeout=10)
    if resp.status_code == 401:
        raise APIError('Token expired')
    if resp.status_code != 200:
        raise APIError(f'API error: {resp.status_code} {resp.text}')
    return resp.json()


def get_last_update(token: str) -> datetime | None:
    """Получить время последнего обновления клиентов."""
    data = api_get('/api/v1/xray/last_update', token)
    updated_at = data.get('updated_at')
    if not updated_at:
        return None
    # Парсим ISO-формат
    return datetime.fromisoformat(updated_at.replace('Z', '+00:00'))


def get_config(token: str) -> dict:
    """Получить конфигурацию XRay."""
    return api_get('/api/v1/xray/config', token)


def write_config(config: dict) -> None:
    """Записать конфигурацию в файл."""
    dir_path = os.path.dirname(XRAY_CONFIG_PATH)
    if dir_path:
        os.makedirs(dir_path, exist_ok=True)
    # Атомарная запись: сначала во временный файл, потом переименуем
    tmp_path = XRAY_CONFIG_PATH + '.tmp'
    with open(tmp_path, 'w') as f:
        json.dump(config, f, indent=2)
        f.write('\n')
    os.replace(tmp_path, XRAY_CONFIG_PATH)
    log.info('Конфигурация записана в %s', XRAY_CONFIG_PATH)


def restart_xray() -> None:
    """Перезапустить XRay-сервер."""
    if not RESTART_COMMAND:
        log.warning('RESTART_COMMAND не задан — перезапуск пропущен')
        return
    log.info('Выполняю перезапуск: %s', RESTART_COMMAND)
    try:
        result = subprocess.run(
            RESTART_COMMAND,
            shell=True,
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode == 0:
            log.info('Перезапуск выполнен успешно')
        else:
            log.error('Перезапуск завершился с кодом %d: %s', result.returncode, result.stderr)
    except subprocess.TimeoutExpired:
        log.error('Таймаут при выполнении команды перезапуска')
    except Exception as e:
        log.error('Ошибка при перезапуске: %s', e)


def run():
    """Основной цикл наблюдателя."""
    log.info('Запуск XRay Config Watcher')
    log.info('API_URL=%s  POLL_INTERVAL=%ds  THRESHOLD=%dmin',
             API_URL, POLL_INTERVAL, THRESHOLD_MINUTES)
    log.info('XRAY_CONFIG_PATH=%s  RESTART_COMMAND=%s',
             XRAY_CONFIG_PATH, RESTART_COMMAND or '(не задана)')

    if not LOGIN_EMAIL or not LOGIN_PASSWORD:
        log.error('LOGIN_EMAIL и LOGIN_PASSWORD должны быть заданы')
        sys.exit(1)

    token = None

    while True:
        try:
            # Авторизация (или реавторизация при просроченном токене)
            if token is None:
                try:
                    token = login()
                except APIError as e:
                    log.error('Ошибка авторизации: %s', e)
                    time.sleep(POLL_INTERVAL)
                    continue

            # Проверяем последнее обновление
            last_update = get_last_update(token)
            now = datetime.now(timezone.utc)
            threshold = now - timedelta(minutes=THRESHOLD_MINUTES)

            if last_update is None:
                log.warning('Не удалось получить время последнего обновления')
            elif last_update > threshold:
                log.info('Обнаружено обновление: last_update=%s',
                         last_update.strftime('%Y-%m-%d %H:%M:%S UTC'))
                # Получаем новую конфигурацию
                try:
                    config = get_config(token)
                    write_config(config)
                    restart_xray()
                except APIError as e:
                    log.error('Ошибка при получении конфига: %s', e)
                    if 'Token expired' in str(e):
                        token = None
            else:
                log.debug('Изменений нет: last_update=%s',
                          last_update.strftime('%Y-%m-%d %H:%M:%S UTC'))

        except APIError as e:
            if 'Token expired' in str(e):
                log.warning('Токен просрочен, реавторизация...')
                token = None
            else:
                log.error('API ошибка: %s', e)
        except requests.ConnectionError:
            log.error('Не удалось подключиться к API (%s)', API_URL)
            token = None  # Возможно, API недоступен
        except Exception as e:
            log.error('Неожиданная ошибка: %s', e, exc_info=True)

        time.sleep(POLL_INTERVAL)


if __name__ == '__main__':
    run()