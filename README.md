# PR Reviewer Assignment Service

Сервис для автоматического назначения ревьюеров на PR из команды автора, с возможностью перераспределения и управления активностью пользователей.

## Возможности

- Управление командами и пользователями
- Автоматическое назначение ревьюеров при создании PR
- Перераспределение ревьюеров
- Массовая деактивация пользователей с автоматическим перераспределением PR
- Статистика

## Быстрый старт

### Локальный запуск

Должны быть установлены
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)
- [Go](https://go.dev/dl/)
- [Goose (CLI для миграций)](https://pressly.github.io/goose/installation/#linux)

**Для MacOS**
```bash
# Запустить контейнер с БД (или можно поднять у себя локально)
docker compose -f docker-compose-db.yml up --build

# Провести миграцию 
GOOSE_DRIVER=postgres \
GOOSE_DBSTRING="postgres://user:pass@localhost:5432/pr_service?sslmode=disable" \
goose -dir ./migrations up

# Запустить программу
go run ./cmd/main.go
```
**Для Windows**
```powershell
# Запустить контейнер с БД (или можно поднять у себя локально)
docker compose -f docker-compose-db.yml up --build

# Миграции (PowerShell)
$env:GOOSE_DRIVER = "postgres"
$env:GOOSE_DBSTRING = "postgres://user:pass@localhost:5432/pr_service?sslmode=disable"

goose -dir ./migrations up

# Миграции (cmd.exe)
set GOOSE_DRIVER=postgres
set GOOSE_DBSTRING=postgres://user:pass@localhost:5432/pr_service?sslmode=disable

goose -dir .\migrations up

# Запустить программу
go run ./cmd/main.go
```
### Docker Compose

```bash
# Запуск всех сервисов
make docker-up
# или docker-compose up

# Остановка
make docker-down
```

## Команды

```bash
make build          # Сборка
make run            # Запуск
make test           # Unit тесты
make test-all       # Все тесты (включая e2e)
make test-cover     # Unit тесты с coverage
make bench          # Бенчмарки
make lint           # Линтинг
make migrate-up     # Применить миграции
make migrate-down   # Откатить миграции
```

## API
 Проверить и использовать API можно двумя способами:
- OpenAPI: `api/openapi.yml`
- Postman: `docs/postman_collection.json`

## Производительность

### DeactivateUsersByTeam

**Apple M4, Go 1.25.4**

| Сценарий           | Пользователи | PR | Время | Память | Аллокации |
|--------------------|--------------|----|-------|--------|-----------|
| Без PR             | 10 | 0 | **13.8 µs** | 10.8 KB | 109 |
| С PR               | 20 | 10 | **141.8 µs** | 102.9 KB | 1,019 |
| Много пользвателей | 100 | 50 | **754.0 µs** | 631.6 KB | 4,759 |

**Другие операции:**

| Операция | Время | Память | Аллокации |
|----------|-------|--------|-----------|
| CreatePR | **21.1 µs** | 14.8 KB | 155 |
| ReassignReviewer | **5.2 µs** | 3.7 KB | 37 |
| MergePR | **8.6 µs** | 6.8 KB | 66 |
| GetTopReviewers | **4.1 µs** | 3.3 KB | 30 |


### Нагрузочное тестирование HTTP

**Инструмент:** `hey` (`go install github.com/rakyll/hey@latest`)

| Нагрузка | Запросов | Конкурентность | Среднее время | P99 | RPS | Успешность |
|----------|----------|----------------|---------------|-----|-----|------------|
| Light | 100 | 10 | **4.7 ms** | 27.6 ms | **2,076** | 100% (200) |
| Medium | 1,000 | 50 | **9.8 ms** | 11.9 ms | **5,001** | 100% (200) |
| Heavy | 5,000 | 100 | **18.1 ms** | 23.4 ms | **5,510** | 100% (200) |


### Основные роуты

**Команды:**
- `POST /team/add` - создание команды
- `POST /team/add-member` - добавление участника
- `GET /team/get?team_name=...` - получение команды

**Пользователи:**
- `POST /users/setIsActive` - изменение активности
- `GET /users/getReview?user_id=...` - PR пользователя
- `POST /users/deactivateByTeam` - деактивация команды

**PR:**
- `POST /pullRequest/create` - создание PR
- `POST /pullRequest/merge` - мерж PR
- `POST /pullRequest/reassign` - перераспределение ревьюера

**Статистика:**
- `GET /stats/prs-total` - общее количество PR
- `GET /stats/prs-status` - PR по статусам
- `GET /stats/top-reviewers` - топ ревьюеров
- `GET /stats/assignments-per-user` - назначения по пользователям
- `GET /stats/avg-close-time` - среднее время закрытия
- `GET /stats/idle-users-per-team` - неактивные пользователи по командам
- `GET /stats/needy-prs-per-team` - PR, требующие ревьюеров

**Health:**
- `GET /health` - проверка работоспособности

## Структура проекта

```
.
├── cmd/              # Точка входа
├── internal/
│   ├── handlers/    # HTTP обработчики
│   ├── services/    # Бизнес-логика
│   ├── repository/  # Слой данных
│   ├── models/      # Модели данных
│   ├── config/      # Конфигурация
│   └── logger/      # Логирование
├── tests/
│   ├── unit/        # Unit тесты
│   ├── integration/ # Интеграционные тесты
│   ├── e2e/         # End-to-end тесты
│   └── benchmark/   # Бенчмарки производительности
├── migrations/      # Миграции БД
├── api/             # OpenAPI спецификация
└── docs/            # Документация
```

## Конфигурация

Переменные окружения (см. `env.example`):

- `PORT` - порт сервера (по умолчанию: 8080)
- `DB_URL` - строка подключения к PostgreSQL
- `LOG_LEVEL` - уровень логирования (debug, info, warn, error)
- `LOG_OUTPUT` - вывод логов (stdout, stderr, file)
- `LOG_FILE_PATH` - путь к файлу логов (если LOG_OUTPUT=file)

## Тестирование

```bash
# Unit тесты
make test

# Все тесты
make test-all

# Тесты с coverage
make test-cover

# Бенчмарки
make bench
```
Тестами покрыто более 77%
## Требования
- Go 1.21+
- PostgreSQL 16+

