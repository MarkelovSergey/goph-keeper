# goph-keeper

Менеджер паролей и учётных данных с архитектурой клиент-сервер.

- **Сервер**: HTTP API на Go с хранилищем PostgreSQL, JWT-аутентификацией и опциональным TLS
- **Клиент**: CLI-приложение для взаимодействия с сервером

## Требования

- Go 1.26+
- Docker и Docker Compose (для локальной базы данных)
- [golangci-lint](https://golangci-lint.run/) (для линтера)

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
```

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
make docker-up       # Запустить контейнер PostgreSQL 16 (порт 5432)
make docker-down     # Остановить и удалить контейнеры
make migrate-up      # Применить миграции
make migrate-down    # Откатить миграции
```

## Документация API

```bash
make swag            # Сгенерировать Swagger-документацию в docs/
```

## Архитектура

```
cmd/
  server/main.go      # Точка входа сервера
  client/main.go      # Точка входа клиента

internal/
  server/
    handler/          # HTTP-обработчики
    middleware/        # HTTP-мидлвары (аутентификация, логирование и т.д.)
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
