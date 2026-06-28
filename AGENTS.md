# AGENT.md — Документация для ИИ-агентов

## Обзор проекта

VPN-менеджмент платформа для администрирования и выдачи доступов к VPN-серверам. Поддерживает два протокола: **XRay VLESS** (с Reality) и **Hysteria2**. Система управляет пользователями, инвайтами (приглашениями) и VPN-клиентами через REST API, с веб-интерфейсом на React и автоматической синхронизацией конфигураций VPN-серверов.

---

## Архитектура

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Frontend   │────▶│   API (Go)   │────▶│  PostgreSQL  │
│  React + TS  │     │  :8081       │     │  :5432       │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────┴───────┐
                     ▼              ▼
              ┌────────────┐ ┌────────────┐
              │xray_server │ │hysteria_server│
              │  (watcher) │ │   (watcher)  │
              └────────────┘ └──────────────┘
```

### Порядок запуска (docker-compose)
1. **db** — PostgreSQL 17
2. **db_init** — применяет `migrations/schema.sql`, завершается
3. **api** — Go API, зависит от db_init
4. **frontend** — Nginx раздаёт React-приложение, проксирует `/api` к API
5. **init_profiles** (закомментирован) — создаёт сервисные учётные записи
"6. **xray_server** (закомментирован) — watcher для синхронизации XRay-конфига
7. **hysteria_server** (закомментирован) — watcher для синхронизации Hysteria-конфига"

---

## Директории

### `api/` — Backend (Go)

REST API на чистом `net/http` с Go 1.22+ route patterns. Порт: `:8081`.

**Структура:**
- `main.go` — точка входа, маршрутизация, panic recovery middleware
- `handlers/` — HTTP-обработчики (thin layer, вызывают middleware-сервисы)
- `middleware/` — бизнес-логика, разделённая по доменам
- `database/` — подключение к PostgreSQL
- `swagger.yaml` — спецификация API (Swagger 2.0)

**Домены middleware:**

| Пакет | Назначение | Сервис (глобальная переменная) |
|---|---|---|
| `security/` | Регистрация, логин, JWT, роли, пользователи | `SecSrv` |
| `xray/` | Управление VLESS-клиентами, генерация XRay-конфига и ссылок | `XraySrv` |
| `hysteria/` | Управление Hysteria2-клиентами, генерация конфига и ссылок | `HystSrv` |
| `invites/` | Создание и активация инвайтов | `InvSrv` |
| `utils/` | JSON-хелперы, CORS | — |

**Паттерн middleware-сервисов:**
Каждый домен имеет структуру-сервис (например, `XrayService`), которая инициализируется как глобальная переменная при старте (`var XraySrv = NewXrayService()`). Сервис содержит `*sql.DB`, параметры подключения и `*log.Logger`. SQL-запросы вынесены в отдельные файлы `sql.go` в каждом пакете.

**Аутентификация:**
- JWT (HS256), токен в заголовке `Authorization: Bearer <token>`
- Срок действия токена: 30 дней
- `SECRET_KEY` из переменной окружения
- `RequireAuth` — middleware для проверки JWT, кладёт `user_id` в context
- `AllowForRole(roleId)` — проверяет, что `access_level` пользователя >= `roleId`

**Роли (access_level):**

| id | role_name | access_level |
|---|---|---|
| 1 | user | 1 |
| 2 | vpn_publisher | 2 |
| 3 | network_admin | 3 |
| 4 | service_admin | 4 |
| 5 | superuser | 5 |

**API-маршруты:**

| Метод | Путь | Доступ | Описание |
|---|---|---|---|
| POST | `/api/v1/security/register` | Публичный | Регистрация |
| POST | `/api/v1/security/login` | Публичный | Логин → JWT |
| GET | `/api/v1/security/me` | Auth | Текущий пользователь |
| GET | `/api/v1/security/users/{user_id}` | Auth | Пользователь по ID |
| POST | `/api/v1/security/users/{user_id}/set_role/{role_id}` | Role 5 | Установить роль |
| GET | `/api/v1/security/roles/{role_id}` | Role 4 | Роль по ID |
| GET | `/api/v1/xray/config` | Role 4 | XRay-конфигурация |
| GET | `/api/v1/xray/get_last_update` | Role 4 | Timestamp последнего обновления |
| GET | `/api/v1/xray/clients/by_user_id/{user_id}` | Auth | VLESS-клиенты пользователя |
| GET | `/api/v1/xray/clients/{client_id}` | Auth | VLESS-клиент по ID |
| GET | `/api/v1/xray/clients/link/{client_id}` | Auth | Ссылка подключения VLESS |
| PATCH | `/api/v1/xray/clients/{client_id}/alias` | Auth | Обновить алиас VLESS-клиента |
| DELETE | `/api/v1/xray/clients/{client_id}` | Auth | Удалить VLESS-клиента (soft delete) |
| GET | `/api/v1/hysteria/config` | Role 4 | Hysteria-конфигурация |
| GET | `/api/v1/hysteria/get_last_update` | Role 4 | Timestamp последнего обновления Hysteria |
| GET | `/api/v1/hysteria/clients/by_user_id/{user_id}` | Auth | Hysteria-клиенты пользователя |
| GET | `/api/v1/hysteria/clients/{client_id}` | Auth | Hysteria-клиент по ID |
| GET | `/api/v1/hysteria/clients/link/{client_id}` | Auth | Ссылка подключения Hysteria |
| PATCH | `/api/v1/hysteria/clients/{client_id}/alias` | Auth | Обновить алиас Hysteria-клиента |
| DELETE | `/api/v1/hysteria/clients/{client_id}` | Auth | Удалить Hysteria-клиента (soft delete) |
| POST | `/api/v1/invite/new` | Role 2 | Создать инвайт |
| POST | `/api/v1/invite/activate` | Auth | Активировать инвайт |

---

### `frontend/` — Frontend (React + TypeScript)

SPA на React 18 + Vite + Tailwind CSS. Nginx раздаёт статику и проксирует `/api` к backend.

**Структура:**
- `src/App.tsx` — маршрутизация, `ProtectedRoute` и `AdminRoute` guards
- `src/context/AuthContext.tsx` — контекст аутентификации, хранит JWT в `localStorage`
- `src/api/client.ts` — typed API-клиент (fetch wrapper)
- `src/types/index.ts` — TypeScript-интерфейсы, зеркалируют API-ответы
- `src/pages/` — страницы (Login, Register, Dashboard, AdminUsers, AdminInvite, InviteActivate)
- `src/components/ConsentSidebar.tsx` — компонент боковой панели
- `nginx.conf` — проксирование `/api/v1/` → `http://api:8081/api/v1/`

**Роуты:**
| Путь | Доступ | Страница |
|---|---|---|
| `/login` | Публичный | LoginPage |
| `/register` | Публичный | RegisterPage |
| `/` | Auth | DashboardPage |
| `/invites/activate` | Auth | InviteActivatePage |
| `/admin/invites` | Role >= 2 | AdminInvitePage |
| `/admin/users` | Role >= 5 | AdminUsersPage |

---

### `migrations/` — Схема БД

`schema.sql` — полная схема PostgreSQL. Применяется контейнером `db_init`.

**Таблицы:**
- `roles` — роли пользователей (id, role_name, access_level)
- `users` — пользователи (id, username, password_hash, email, role_id, status)
- `invites` — инвайты (id, invite_hash, expires_at, status, vpn_type)
- `vless_clients` — VLESS-клиенты (id, access_key UUID, user_id, invite_id, alias, status)
- `hysteria_clients` — Hysteria-клиенты (id, password UUID, user_id, invite_id, alias, status)

**Связи:**
- `users.role_id` → `roles.id`
- `vless_clients.user_id` → `users.id` (CASCADE DELETE)
- `vless_clients.invite_id` → `invites.id` (CASCADE DELETE, UNIQUE)
- `hysteria_clients.user_id` → `users.id` (CASCADE DELETE)
- `hysteria_clients.invite_id` → `invites.id` (CASCADE DELETE, UNIQUE)

**Инвайты:** статус может быть `active`, `activated`, `expired`, `revoked`. Поле `vpn_type` определяет, какой тип клиента будет создан при активации (`xray` или `hysteria`).

---

### `init_profiles/` — Инициализация сервисных аккаунтов

Python-скрипт (`init_profiles.py`), который создаёт служебных пользователей при первом запуске:
- **watcher** (role_id=4, service_admin) — для xray_server watcher
- **superuser** (role_id=5, superuser) — главный администратор

Пароли хэшируются bcrypt (rounds=10, совместимо с Go `bcrypt.GenerateFromPassword`). Используется `ON CONFLICT DO NOTHING` — безопасно при повторных запусках.

В `docker-compose.yml` сервис закомментирован.

---

### `xray_server/` — Watcher для XRay

Python-скрипт (`watcher.py`), который периодически опрашивает API и поддерживает XRay-конфиг в актуальном состоянии.

**Логика работы:**
1. Логинится в API, получает JWT
2. С заданным интервалом проверяет `GET /api/v1/xray/get_last_update`
3. Если timestamp изменился — скачивает конфиг `GET /api/v1/xray/config`
4. Валидирует конфиг (`xray run -test -c <path>`)
5. Атомарно записывает конфиг (временный файл → `os.replace`)
6. Выполняет `RESTART_COMMAND` для перезапуска XRay (обычно kill процесса, `start.sh` поднимет его заново)

**Дополнительно:** `start.sh` — shell-скрипт для запуска XRay в контейнере.

В `docker-compose.yml` сервис закомментирован.

---

### `hysteria_server/` — Watcher для Hysteria2
Python-скрипт (`watcher.py`), который периодически опрашивает API и поддерживает Hysteria2-конфиг в актуальном состоянии. Аналогичен `xray_server/`, но адаптирован для Hysteria2.
**Логика работы:**
1. Логинится в API, получает JWT
2. С заданным интервалом проверяет `GET /api/v1/hysteria/get_last_update`
3. Если timestamp изменился — скачивает конфиг `GET /api/v1/hysteria/config`
4. Валидирует конфиг (`hysteria check -c <path>`)
5. Атомарно записывает конфиг в JSON (временный файл → `os.replace`)
6. Выполняет `RESTART_COMMAND` для перезапуска Hysteria2 (обычно kill процесса, `start.sh` поднимет его заново)
   **Структура:**
- `watcher.py` — основной watcher-скрипт
- `start.sh` — shell-скрипт для запуска Hysteria2 в контейнере (запускает watcher, ждёт конфиг, запускает `hysteria server`, авторестарт при падении)
- `Dockerfile` — Python 3.12 Alpine + Hysteria2 binary из GitHub Releases
- `requirements.txt` — `requests`
  **Отличия от xray_server:**
- Бинарник `hysteria` вместо `xray`
- Валидация: `hysteria check -c <path>` вместо `xray run -test -c <path>`
- Запуск сервера: `hysteria server -c <path>` вместо `xray run -c <path>`
- Конфиг в JSON (как отдаёт API), путь по умолчанию `/etc/hysteria/config.json`
- Эндпоинты: `/api/v1/hysteria/get_last_update` и `/api/v1/hysteria/config`
- Порт 643 (TCP + UDP)
  В `docker-compose.yml` сервис закомментирован.
---

## Ключевые концепции

### Инвайт-система
1. Пользователь с ролью >= 2 (vpn_publisher) создаёт инвайт: `POST /api/v1/invite/new` с `invite_word` и `vpn_type` (`vless` или `hysteria`)
2. Другой пользователь активирует инвайт: `POST /api/v1/invite/activate` с `invite_word` и `alias`
3. При активации создаётся запись в `vless_clients` или `hysteria_clients` (зависит от `vpn_type` инвайта)
4. `access_key` (VLESS) и `password` (Hysteria) генерируются автоматически как UUID

### Генерация ссылок подключения
- **VLESS:** `xray/link.go` — `makeXrayLinkParams()` формирует ссылку на основе параметров XRay (Reality) из переменных окружения
- **Hysteria:** `hysteria/link.go` — `makeXrayLinkParams()` формирует ссылку hysteria2:// с параметрами (obfs=salamander, ACME-домен, порт)

### Генерация серверных конфигураций
- **XRay:** `xray/xray_config.go` — `GetConfig()` собирает полный XRay JSON-конфиг, включая всех активных VLESS-клиентов из БД
- **Hysteria:** `hysteria/hysteria_config.go` — `GetConfig()` собирает Hysteria2 YAML-конфиг с userpass-аутентификацией на основе всех активных клиентов из БД

---

## Переменные окружения

| Переменная | Назначение | Используется в |
|---|---|---|
| `POSTGRES_PASSWORD` | Пароль PostgreSQL | db, db_init, api, init_profiles |
| `POSTGRES_HOST` | Хост PostgreSQL | api |
| `POSTGRES_PORT` | Порт PostgreSQL | api |
| `SECRET_KEY` | Секрет для JWT | api |
| `XRAY_HOST` | Хост XRay-сервера | api |
| `XRAY_CONFIG_TOKEN` | Токен конфига XRay | api |
| `XRAY_LISTEN` | Адрес прослушивания XRay | api |
| `XRAY_PORT` | Порт XRay | api |
| `XRAY_DEST` | Dest для Reality | api |
| `XRAY_SERVER_NAMES` | Server names (через запятую) | api |
| `XRAY_PRIVATE_KEY` | Приватный ключ Reality | api |
| `XRAY_SHORT_IDS` | Short IDs (через запятую) | api |
| `HYSTERIA_LISTEN` | Адрес прослушивания Hysteria | api |
| `HYSTERIA_ACME_EMAIL` | Email для ACME | api |
| `ACME_DOMAINS` | Домены ACME (через запятую) | api |
| `HYSTERIA_OBFS_PASSWORD` | Пароль обфускации (salamander) | api |
| `HYSTERIA_MASQUERADE_PROXY_URL` | URL маскарад-прокси | api |
| `HYSTERIA_BANDWIDTH_UP` | Пропускная способность вверх (mbps) | api |
| `HYSTERIA_BANDWIDTH_DOWN` | Пропускная способность вниз (mbps) | api |
| `WATCHER_*` | Параметры watcher/xray_server | xray_server, init_profiles |
| `WATCHER_HYSTERIA_EMAIL` | Email для логина watcher Hysteria | hysteria_server |
| `WATCHER_HYSTERIA_PASSWORD` | Пароль для логина watcher Hysteria | hysteria_server |
| `WATCHER_HYSTERIA_CONFIG_PATH` | Путь к конфигу Hysteria | hysteria_server |

---

## Конвенции кода

### Go (api/)
- Стандартная библиотека `net/http`, без фреймворков
- Go 1.22+ route patterns (`GET /path/{param}`)
- Сервисы инициализируются как глобальные переменные (`var XraySrv = NewXrayService()`)
- SQL-запросы в отдельных файлах `sql.go`
- Обработка ошибок: кастомные ошибки (`var ErrorClientNotFound = errors.New(...)`) + `errors.Is()`
- JSON-ответы через утилиты `utils.WriteJSON` / `utils.ReadJSON`
- Panic recovery в `main.go` (`withRecover`)

### TypeScript (frontend/)
- Vite + React 18 + Tailwind CSS
- Путь `@/` → `src/`
- Типы в `src/types/index.ts`, зеркалируют API-ответы
- API-клиент в `src/api/client.ts` с автоматической подстановкой JWT
- Auth-контекст через React Context (`useAuth()`)

### Python (init_profiles/, xray_server/)
- Скрипты утилитарного характера
- `psycopg2` для БД, `requests` для HTTP, `bcrypt` для хэширования
- Переменные окружения через `os.getenv()`