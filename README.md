# goph-keeper

Менеджер паролей и учётных данных с архитектурой клиент-сервер.

- **Сервер**: HTTP API на Go с хранилищем PostgreSQL, JWT-аутентификацией и опциональным TLS
- **Клиент**: CLI-приложение для взаимодействия с сервером

## Требования

- **Go 1.26.1+** — язык разработки
- **Docker и Docker Compose** — для запуска локальной базы данных PostgreSQL 16
- **GNU Make** — для запуска команд из `Makefile`
- **[golang-migrate](https://github.com/golang-migrate/migrate)** (`migrate`) — применение миграций базы данных
- **[golangci-lint](https://golangci-lint.run/)** — статический анализ кода
- **[swaggo/swag](https://github.com/swaggo/swag)** (`swag`) — генерация документации Swagger (опционально)

## Быстрый старт

```bash
# Скопировать конфигурацию
cp .env.example .env

# Запустить PostgreSQL
make docker-up

# Применить миграции
make migrate-up

# Запустить сервер
make run-server

# В другом терминале — запустить клиент
make run-client

# Запустить клиент с командой
make run-client ARGS="version"
make run-client ARGS="add -h"
```

## Запуск

```bash
make run-server                    # Запустить сервер
make run-client                    # Запустить клиент (интерактивно)
make run-client ARGS="version"     # Передать команду клиенту
make run-client ARGS="add -h"      # Передать команду с флагами
make run-client version            # Краткая форма (trailing target)
```

> Для передачи аргументов используйте переменную `ARGS` или просто допишите
> команду после цели (trailing targets). Аналогично для `run-server ARGS="..."`.

## Сборка

```bash
make build           # Собрать сервер и клиент в bin/
make build-server    # bin/goph-keeper-server
make build-client    # bin/goph-keeper
make build-all       # Кросс-платформенная сборка (Linux, macOS, Windows)
```

## Тесты и линтер

```bash
make test            # Запустить все тесты
make test-cover      # Тесты с отчётом о покрытии (coverage.out)
make lint            # Запустить golangci-lint
```

## Переменные окружения

Скопируйте `.env.example` в `.env`. Загрузчик конфигурации ищет `.env` вверх по дереву директорий.

### Сервер

| Переменная     | По умолчанию | Описание                                      |
| -------------- | ------------ | --------------------------------------------- |
| `LISTEN_ADDR`  | `:8080`      | TCP-адрес для прослушивания                   |
| `DATABASE_DSN` | —            | Строка подключения к PostgreSQL (обязательно) |
| `JWT_SECRET`   | —            | HMAC-секрет для JWT-токенов (обязательно)     |
| `TLS_CERT`     | —            | Путь к TLS-сертификату (необязательно)        |
| `TLS_KEY`      | —            | Путь к TLS-ключу (необязательно)              |

### Клиент

| Переменная              | По умолчанию            | Описание                                  |
| ----------------------- | ----------------------- | ----------------------------------------- |
| `SERVER_ADDRESS`        | `http://localhost:8080` | Базовый URL сервера                       |
| `TLS_INSECURE`          | `false`                 | Отключить проверку TLS-сертификата        |
| `GOPHKEEPER_CONFIG_DIR` | `~/.gophkeeper`         | Директория для хранения состояния клиента |

## База данных

```bash
make docker-up   # Запустить контейнер PostgreSQL 16 (порт 5432)
make docker-down # Остановить и удалить контейнеры
```

Миграции требуют явной передачи строки подключения:

```bash
make migrate-up   DATABASE_DSN="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
make migrate-down DATABASE_DSN="postgres://user:pass@localhost:5432/dbname?sslmode=disable"
```

## Очистка

```bash
make clean           # Удалить bin/ и coverage.out
```

## Документация API

После запуска сервера Swagger UI доступен по адресу:

```
http://localhost:8080/swagger/index.html
```

> Перед открытием выполните `make swag` для генерации документации, затем перезапустите сервер.

## Postman-коллекция

Файл `goph-keeper.postman_collection.json` содержит готовые запросы для ручного тестирования API.

**Импорт:** Postman → Import → выбрать файл коллекции.

**Переменные коллекции:**

| Переменная      | Описание                                                  |
| --------------- | --------------------------------------------------------- |
| `base_url`      | Базовый URL сервера (по умолчанию `http://localhost:8080`) |
| `token`         | JWT-токен, сохраняется автоматически после входа/регистрации |
| `credential_id` | UUID последней созданной записи, сохраняется автоматически |

**Состав коллекции:**

- **Аутентификация**
  - `POST /api/register` — регистрация нового пользователя
  - `POST /api/login` — вход существующего пользователя
- **Учётные данные** (все запросы требуют токен)
  - `GET /api/credentials` — список всех записей
  - `POST /api/credentials` — создать запись (типы: `login_password`, `text`, `binary`, `bank_card`)
  - `GET /api/credentials/{id}` — получить запись по UUID
  - `PUT /api/credentials/{id}` — обновить запись
  - `DELETE /api/credentials/{id}` — удалить запись

**Порядок работы:**
1. Запустить сервер: `make run-server`
2. Выполнить «Регистрация» или «Вход» — токен сохранится в переменную `token`
3. Выполнять запросы к учётным данным

> Поле `data` содержит зашифрованный blob в кодировке base64. Клиент шифрует данные перед отправкой (AES-256-GCM), сервер хранит непрозрачный blob без расшифровки.

## Архитектура

```
cmd/
  server/main.go      # Точка входа сервера
  client/main.go      # Точка входа клиента

internal/
  server/
    handler/          # HTTP-обработчики
    middleware/       # HTTP-мидлвары (аутентификация, логирование и т.д.)
    service/          # Бизнес-логика
    repository/       # Доступ к данным (PostgreSQL)
    model/            # Модели данных сервера
    app/              # Инициализация приложения
    config/           # Конфигурация
  client/
    cmd/              # Определения команд Cobra
    api/              # HTTP-клиент для общения с сервером
    app/              # Логика клиентского приложения
    crypto/           # Криптографические операции
    config/           # Конфигурация
  model/              # Общие модели (сервер + клиент)
```
