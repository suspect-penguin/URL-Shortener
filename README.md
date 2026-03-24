# URL Shortener

HTTP-сервис для сокращения ссылок. Написан на Go потому что было интерено понять как оно работает, а лучше всего понимаешь когда сам делаешь.

## Стек

- **Go 1.22** — стандартная библиотека + chi router
- **PostgreSQL 16** — хранение ссылок
- **In-memory cache** — горутин инвалидации
- **Docker / Docker Compose** — запуск и сборка образа

## Архитектура

```
cmd/server/main.go          —  входик, graceful shutdown
internal/
  config/                   — конфиг
  model/                    — URL
  cache/                    — потокобезопастный кэш
  repository/               — интерфейс и pgx
  service/                  — логика
  handler/                  — хендлеры
migrations/001_init.sql     — схема
frontend/index.html         — красотень
```

## API

| Метод | Описание |
|-------|----------|
| `POST` | Создать короткую ссылку |
| `GET` | Редирект на оригинальную ссылку |
| `GET` | Статистика по ссылке |


## Запуск


```bash
docker compose up --build
```

Сервис на [http://localhost:8080](http://localhost:8080)

---

### Локально

**Требования:** Go 1.22+, PostgreSQL

```bash

createdb urlshortener

psql -d urlshortener -f migrations/001_init.sql

cp .env.example .env

go run ./cmd/server
```

Переменные окружения:

| Переменная | Дефолт |
|-----------|--------|
| `HTTP_ADDR` | `:8080` |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/urlshortener?sslmode=disable` |
| `BASE_URL` | `http://localhost:8080` |

---



## Тесты

```bash
go test ./...

go test -race ./...
```


## CI


- **test** — `go vet` + `go test -race` с PostgreSQL
- **lint** — `golangci-lint`
- **build** — сборка и публикация Docker-образа на Docker Hub
