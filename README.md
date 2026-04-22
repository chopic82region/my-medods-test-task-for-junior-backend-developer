# Task Service

Сервис для управления задачами с HTTP API на Go.

## Требования

- Go `1.23+`
- Docker и Docker Compose

## Быстрый запуск через Docker Compose

```bash
docker compose up --build
```

После запуска сервис будет доступен по адресу `http://localhost:8081`.

Если `postgres` уже запускался ранее со старой схемой, пересоздай volume:

```bash
docker compose down -v
docker compose up --build
```

Причина в том, что SQL-файл из `migrations/0001_create_tasks.up.sql` монтируется в `docker-entrypoint-initdb.d` и применяется только при инициализации пустого data volume.

## Swagger

Swagger UI:

```text
http://localhost:8081/swagger/
```

OpenAPI JSON:

```text
http://localhost:8081/swagger/openapi.json
```

## API

Базовый префикс API:

```text
/api/v1
```

Основные маршруты:

- `POST /api/v1/tasks`
- `GET /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PUT /api/v1/tasks/{id}`
- `DELETE /api/v1/tasks/{id}`

Добавлено: 

# 🔄 Поддержка периодических задач:

-daily	Ежедневные задачи (каждый n-й день)	{"interval": N}
    
-monthly	Ежемесячные задачи на определенные числа	{"days": [5, 20]}
    
-specific_dates	Задачи на конкретные даты	{"dates": ["2026-05-01", "2026-05-15"]}

-even_odd	Задачи на четные/нечетные дни	{"parity": "even"} или "odd"

# Примеры

# Создание ежедневной задачи:

curl -X POST http://localhost:8080/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d '{
        "title": "Ежедневный обзвон пациентов",
        "description": "Обзванивать пациентов по списку",
        "status": "new",
        "recurrence_type": "daily",
        "recurrence_config": {
             "interval": 1
    }
  }'

# Создание задачи с переодичностью:

curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Полив цветов",
    "description": "Поливать комнатные растения",
    "status": "new",
    "recurrence_type": "daily",
    "recurrence_config": {
      "interval": 3
    }
  }'

# Создание ежемесячной задачи:

curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Формирование отчета",
    "description": "Сформировать ежемесячный отчет по продажам",
    "status": "new",
    "recurrence_type": "monthly",
    "recurrence_config": {
      "days": [5, 20]
    }
  }'

# Создание ежемесячной задачи (5-го и 20-го числа):

curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Формирование отчета",
    "description": "Сформировать ежемесячный отчет по продажам",
    "status": "new",
    "recurrence_type": "monthly",
    "recurrence_config": {
      "days": [5, 20]
    }
  }'

# Создание задачи на конкретные даты:

curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Встреча с клиентом",
    "description": "Провести переговоры",
    "status": "new",
    "recurrence_type": "specific_dates",
    "recurrence_config": {
      "dates": ["2026-05-01", "2026-05-15", "2026-05-30"]
    }
  }'

# Новые миграции: 

-- Добавление полей для периодичности
ALTER TABLE tasks ADD COLUMN recurrence_type VARCHAR(20) NOT NULL DEFAULT 'none';
ALTER TABLE tasks ADD COLUMN recurrence_config JSONB;
ALTER TABLE tasks ADD COLUMN parent_task_id BIGINT REFERENCES tasks(id) ON DELETE CASCADE;

-- Комментарии к полям
COMMENT ON COLUMN tasks.recurrence_type IS 'Тип периодичности: none, daily, monthly, specific_dates, even_odd';
COMMENT ON COLUMN tasks.recurrence_config IS 'JSON с параметрами периодичности';
COMMENT ON COLUMN tasks.parent_task_id IS 'ID шаблона-родителя для периодических задач';

# 🧪 Тестирование
Юнит-тесты
В проекте реализованы юнит-тесты для проверки логики:

# Запуск всех тестов
    go test ./...

# Запуск с детальным выводом
    go test -v ./...

# Запуск с покрытием
    go test -cover ./...