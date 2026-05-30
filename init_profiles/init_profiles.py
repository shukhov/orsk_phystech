#!/usr/bin/env python3
"""
Создаёт сервисные учётные записи в БД при первом запуске.
Пароли хэшируются bcrypt (как в Go API: bcrypt.GenerateFromPassword, DefaultCost).
ON CONFLICT DO NOTHING — безопасно при повторных запусках.
"""

import os
import sys
import logging
import bcrypt
import psycopg2

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s [%(levelname)s] %(message)s',
    datefmt='%Y-%m-%d %H:%M:%S',
)
log = logging.getLogger('init_profiles')

# --- Переменные окружения ---
DB_HOST = os.environ.get('DB_HOST', 'db')
DB_PORT = int(os.environ.get('DB_PORT', '5432'))
DB_NAME = os.environ.get('DB_NAME', 'vpn_db')
DB_USER = os.environ.get('DB_USER', 'vpn_api')
DB_PASSWORD = os.environ.get('DB_PASSWORD', '')

WATCHER_USERNAME = os.environ.get('WATCHER_USERNAME', '')
WATCHER_EMAIL = os.environ.get('WATCHER_EMAIL', '')
WATCHER_PASSWORD = os.environ.get('WATCHER_PASSWORD', '')

SUPERUSER_USERNAME = os.environ.get('SUPERUSER_USERNAME', '')
SUPERUSER_EMAIL = os.environ.get('SUPERUSER_EMAIL', '')
SUPERUSER_PASSWORD = os.environ.get('SUPERUSER_PASSWORD', '')

UPSERT_QUERY = """
INSERT INTO public.users (username, password_hash, email, role_id)
VALUES (%s, %s, %s, %s)
ON CONFLICT (email) DO NOTHING
"""


def hash_password(password: str) -> str:
    """Хэширует пароль bcrypt. Совместимо с Go bcrypt.GenerateFromPassword."""
    salt = bcrypt.gensalt(rounds=10)  # bcrypt.DefaultCost в Go = 10
    hashed = bcrypt.hashpw(password.encode('utf-8'), salt)
    return hashed.decode('utf-8')


def create_user(cursor, username: str, email: str, password: str, role_id: int) -> bool:
    """Создаёт пользователя, если его ещё нет."""
    if not username or not email or not password:
        log.warning('Пропуск: не все данные заданы для пользователя (username=%s, email=%s)',
                    username, email)
        return False
    password_hash = hash_password(password)
    cursor.execute(UPSERT_QUERY, (username, password_hash, email, role_id))
    if cursor.rowcount > 0:
        log.info('Создан пользователь: %s (%s), role_id=%d', username, email, role_id)
        return True
    else:
        log.info('Пользователь уже существует: %s (%s)', username, email)
        return False


def main():
    log.info('Подключение к БД: %s@%s:%s/%s', DB_USER, DB_HOST, DB_PORT, DB_NAME)
    try:
        conn = psycopg2.connect(
            host=DB_HOST,
            port=DB_PORT,
            dbname=DB_NAME,
            user=DB_USER,
            password=DB_PASSWORD,
        )
    except Exception as e:
        log.error('Не удалось подключиться к БД: %s', e)
        sys.exit(1)

    try:
        with conn:
            with conn.cursor() as cur:
                create_user(cur, WATCHER_USERNAME, WATCHER_EMAIL, WATCHER_PASSWORD, role_id=4)
                create_user(cur, SUPERUSER_USERNAME, SUPERUSER_EMAIL, SUPERUSER_PASSWORD, role_id=5)
    except Exception as e:
        log.error('Ошибка при создании пользователей: %s', e)
        sys.exit(1)
    finally:
        conn.close()

    log.info('Готово')


if __name__ == '__main__':
    main()