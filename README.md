# URL Shortener

URL Shortener — это сервис для сокращения ссылок, который поддерживает два типа хранилищ: in-memory и PostgreSQL. Он предоставляет интерфейс через gRPC и может быть развернут в Docker.

## Особенности

- **Поддержка двух типов хранилищ**: in-memory и PostgreSQL.
- **gRPC интерфейс** для взаимодействия с сервисом.
- **Docker**-образ для простого развертывания.
- Легко настраиваемый через переменные окружения.

## Установка и запуск

### 1. Склонируйте репозиторий

```bash
git clone https://github.com/ivaeg3/url-shortener.git
cd url-shortener
```

### 2. Соберите Docker-образ

Для того, чтобы собрать Docker-образ и запустить сервис, используйте следующую команду:

```bash
docker build -t url-shortener .
```

### 3. Запуск в Docker

```bash
docker run -d -p 50051:50051 \
    -e STORAGE_TYPE=memory \
    -e PORT=50051 \
    url-shortener
```

### 4. Запуск локально

Установите зависимости:

```bash
go mod tidy
```

Скомпилируйте и запустите сервер:

```bash
go run cmd/server/main.go
```

### 5. Конфигурация

Для настройки сервиса используйте переменные окружения или опции:

- `PORT` (`-port`) — Порт для gRPC сервера (по умолчанию `50051`).
- `STORAGE_TYPE` (`-storage-type`) — Тип хранилища: `memory` или `postgres` (по умолчанию `memory`).
- `POSTGRES_URL` (`-postgres-url`) — URL для подключения к базе данных PostgreSQL (требуется, если `STORAGE_TYPE=postgres`).

Например:

```bash
export STORAGE_TYPE=postgres
export POSTGRES_URL="postgres://user:password@localhost:5432/dbname"
```

Или:
```bash
go run cmd/server/main.go -port 50051 -storage-type postgres -postgres-url "postgres://user:password@localhost:5432/dbname"
```