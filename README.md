# Task Manager

HTTP API для управления задачами. Сервер написан на Go с использованием стандартного пакета `net/http`. Задачи хранятся в PostgreSQL 17, для работы с базой используется `pgxpool`. Приложение и база запускаются через Docker Compose.

## Запуск

Понадобится запущенный Docker с плагином Compose. Устанавливать Go локально для этого способа запуска не нужно: приложение собирается внутри Docker.

Из корня проекта выполни:

```bash
docker compose up -d --build
```

API доступен по адресу `http://127.0.0.1:8080`. Проверить запуск:

```bash
docker compose ps
curl -i http://127.0.0.1:8080/tasks
```

Для остановки используй `docker compose down`. Задачи сохраняются в именованном томе `postgres_data`. Команда `docker compose down -v` удаляет том вместе с данными.

## API

| Метод и путь | Действие | Успешный ответ |
| --- | --- | --- |
| `GET /tasks` | Получить список задач, отсортированный по ID | `200 OK`, объект с полем `tasks` |
| `GET /tasks?done=true` | Показать только выполненные задачи; для невыполненных используй `done=false` | `200 OK` |
| `POST /tasks` | Создать задачу с полем `title` | `201 Created`, объект с ID задачи |
| `PATCH /tasks/{id}` | Изменить название или отметить задачу выполненной | `204 No Content` |
| `DELETE /tasks/{id}` | Удалить задачу по ID | `204 No Content` |

### Примеры запросов

Создать задачу:

```bash
curl -i -X POST http://127.0.0.1:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Изучить Go"}'
```

Пример ответа: `{"id":1}`. В командах ниже замени `1` на полученный ID.

```bash
# Отметить задачу выполненной
curl -i -X PATCH http://127.0.0.1:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"done":true}'

# Показать выполненные задачи
curl -i 'http://127.0.0.1:8080/tasks?done=true'

# Удалить задачу
curl -i -X DELETE http://127.0.0.1:8080/tasks/1
```

## Подключение к PostgreSQL

Приложение получает строку подключения из `DATABASE_URL`. В сети Compose PostgreSQL доступен по адресу `db:5432`.

Чтобы открыть БД из GoLand или другого клиента на своём компьютере, укажи:

| Поле | Значение |
| --- | --- |
| Host | `127.0.0.1` |
| Port | `15432` |
| Database | `taskapp` |
| User | `taskapp` |
| Password | Пароль пользователя PostgreSQL |

При создании БД на пустом томе PostgreSQL берёт пароль из `POSTGRES_PASSWORD` и выполняет [init.sql](init.sql), создавая таблицу `tasks`. На уже инициализированном томе изменение этих настроек не меняет пароль и не применяет SQL повторно.

## Структура проекта

| Файл или каталог | Назначение |
| --- | --- |
| [main.go](main.go) | Подключение к БД, регистрация маршрутов и запуск HTTP-сервера |
| [handlers/](handlers/) | Обработка HTTP-запросов, проверка входных данных и ответы в JSON |
| [repository/](repository/) | SQL-запросы к PostgreSQL |
| [tasks/](tasks/) | Структура задачи |
| [init.sql](init.sql) | Создание таблицы при инициализации базы |
| [Dockerfile](Dockerfile), [docker-compose.yaml](docker-compose.yaml) | Сборка приложения и запуск контейнеров |
