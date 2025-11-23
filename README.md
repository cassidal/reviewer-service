# Reviewer Service

Сервис для управления назначением ревьюверов на Pull Request'ы.

## Описание

Сервис предоставляет REST API для:
- Управления командами и пользователями
- Создания Pull Request'ов с автоматическим назначением ревьюверов
- Управления статусами PR (merge)
- Переназначения ревьюверов

## Требования

- Go 1.25+
- PostgreSQL 16+
- Docker и Docker Compose (для запуска через Docker)

## Быстрый старт

### Вариант 1: Запуск через Docker Compose (рекомендуется)

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd reviewer-service
```

2. Запустите все сервисы:
```bash
make docker-up
```

Это команда:
- Запустит PostgreSQL базу данных
- Применит миграции
- Запустит backend сервис

3. Проверьте, что сервис работает:
```bash
curl http://localhost:8080/team/get?team_name=test
```

### Вариант 2: Локальный запуск

1. Установите зависимости:
```bash
make deps
```

2. Настройте PostgreSQL базу данных:
```bash
createdb reviewer_db
```

3. Примените миграции:
```bash
# Используя migrate CLI
migrate -path ./migrations -database "postgres://user:password@localhost:5432/reviewer_db?sslmode=disable" up

# Или через Docker
make migrate-up
```

4. Настройте конфигурацию:
Отредактируйте `config/local.yaml` с вашими настройками БД.

5. Соберите и запустите:
```bash
make build
make run
```

## Структура проекта

```
reviewer-service/
├── cmd/
│   └── reviewer-service/     # Точка входа приложения
├── internal/
│   ├── config/                # Конфигурация
│   ├── domain/                # Бизнес-логика (domain layer)
│   │   ├── team/
│   │   ├── user/
│   │   └── pullrequest/
│   ├── http-server/           # HTTP handlers
│   ├── storage/               # Репозитории (infrastructure layer)
│   └── tests/                # Тесты
├── migrations/                # SQL миграции
├── config/                   # Конфигурационные файлы
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

## API Endpoints

### Teams

#### POST /team/add
Создать команду с участниками.

**Request:**
```json
{
  "team": {
    "team_name": "backend",
    "members": [
      {
        "user_id": "u1",
        "username": "Alice",
        "is_active": true
      }
    ]
  }
}
```

**Response:** `201 Created`
```json
{
  "team": {
    "team_name": "backend",
    "members": ["..."]
  }
}
```

#### GET /team/get?team_name=backend
Получить команду с участниками.

**Response:** `200 OK`
```json
{
  "team_name": "backend",
  "members": ["..."]
}
```

### Users

#### POST /users/setIsActive
Установить флаг активности пользователя.

**Request:**
```json
{
  "user_id": "u1",
  "is_active": false
}
```

**Response:** `200 OK`
```json
{
  "user": {
    "user_id": "u1",
    "username": "Alice",
    "team_name": "backend",
    "is_active": false
  }
}
```

#### GET /users/getReview?user_id=u1
Получить PR'ы, где пользователь назначен ревьювером.

**Response:** `200 OK`
```json
{
  "user_id": "u1",
  "pull_requests": ["..."]
}
```

### Pull Requests

#### POST /pullRequest/create
Создать PR и автоматически назначить до 2 ревьюверов из команды автора.

**Request:**
```json
{
  "pull_request_id": "pr-1",
  "pull_request_name": "Add feature",
  "author_id": "u1"
}
```

**Response:** `201 Created`
```json
{
  "pr": {
    "pull_request_id": "pr-1",
    "pull_request_name": "Add feature",
    "author_id": "u1",
    "status": "OPEN",
    "assigned_reviewers": ["u2", "u3"]
  }
}
```

#### POST /pullRequest/merge
Пометить PR как MERGED (идемпотентная операция).

**Request:**
```json
{
  "pull_request_id": "pr-1"
}
```

**Response:** `200 OK`
```json
{
  "pr": {
    "pull_request_id": "pr-1",
    "status": "MERGED",
    "mergedAt": "2025-10-24T12:34:56Z"
  }
}
```

#### POST /pullRequest/reassign
Переназначить конкретного ревьювера на другого из его команды.

**Request:**
```json
{
  "pull_request_id": "pr-1",
  "old_reviewer_id": "u2"
}
```

**Response:** `200 OK`
```json
{
  "pr": {
    "pull_request_id": "pr-1",
    "assigned_reviewers": ["u3", "u5"]
  },
  "replaced_by": "u5"
}
```

## Команды Makefile

```bash
make help              # Показать справку по всем командам
make build             # Собрать проект
make run               # Запустить приложение локально
make test              # Запустить тесты
make test-integration  # Запустить интеграционные тесты
make clean             # Очистить скомпилированные файлы
make docker-build      # Собрать Docker образ
make docker-up         # Запустить все сервисы в Docker
make docker-down       # Остановить все сервисы
make docker-logs       # Показать логи
make migrate-up        # Применить миграции
make migrate-down      # Откатить миграции
make deps              # Установить зависимости
make fmt               # Форматировать код
make vet               # Запустить go vet
```

## Миграции

Миграции находятся в директории `migrations/`:
- `000_initial_schema.sql` - создание таблиц team и users
- `001_create_pull_requests.sql` - создание таблиц pull_requests и pr_reviewers

Для применения миграций через Docker:
```bash
make migrate-up
```

Для отката:
```bash
make migrate-down
```

## Тестирование

### Запуск всех тестов:
```bash
make test
```

### Запуск интеграционных тестов:
```bash
make test-integration
```

Интеграционные тесты используют testcontainers-go для автоматического запуска PostgreSQL контейнера. Внешняя база данных не требуется - контейнер поднимается и останавливается автоматически для каждого теста.

**Требования:** Docker должен быть установлен и запущен, так как testcontainers использует Docker для запуска контейнеров.

## Конфигурация

Конфигурация приложения задается через YAML файл, путь к которому указывается в переменной окружения `CONFIG_PATH`.

Пример конфигурации (`config/local.yaml`):
```yaml
env: dev
datasource:
  host: localhost
  port: 5432
  database: reviewer_db
  username: reviewer
  password: reviewer_password
  timeout: 5s
http_server:
  host: 0.0.0.0
  port: 8080
  timeout: 4s
  idle_timeout: 30s
```

## Docker

### Сборка образа:
```bash
make docker-build
```

### Запуск всех сервисов:
```bash
make docker-up
```

### Просмотр логов:
```bash
make docker-logs
```

### Остановка:
```bash
make docker-down
```

## Архитектура

Проект следует принципам Clean Architecture:
- **Domain layer** (`internal/domain/`) - бизнес-логика, не зависит от внешних слоев
- **Infrastructure layer** (`internal/storage/`) - реализация репозиториев для PostgreSQL
- **Presentation layer** (`internal/http-server/`) - HTTP handlers

## Разработка

### Добавление новой миграции:
```bash
make migrate-create NAME=add_new_table
```

### Форматирование кода:
```bash
make fmt
```

### Проверка кода:
```bash
make vet
```

## Лицензия

MIT

