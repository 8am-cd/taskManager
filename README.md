# Task Manager

Task Manager — веб-приложение для управления личными задачами с HTTP API на Go (net/http). Пользователи могут зарегистрироваться, войти в аккаунт и работать только со своими задачами. Авторизация основана на серверных сессиях: браузер передаёт идентификатор сессии в cookie. Аккаунты, сессии и задачи хранятся в PostgreSQL 17, к которому приложение подключается через pgxpool. Приложение и база данных запускаются через Docker Compose.

## Запуск

Понадобится запущенный Docker с плагином Compose. Устанавливать Go локально для этого способа запуска не нужно: приложение собирается внутри Docker.

Из корня проекта выполни:

```bash
docker compose up -d --build
```

Открой [веб-интерфейс](http://127.0.0.1:8080/) в браузере, зарегистрируйся и войди. В аккаунте можно создавать, переименовывать и удалять свои задачи, отмечать их выполненными и фильтровать список. API доступен на том же порту. Проверить запуск:

```bash
docker compose ps
curl -i http://127.0.0.1:8080/
```

Для остановки используй `docker compose down`. Задачи сохраняются в именованном томе `postgres_data`. Команда `docker compose down -v` удаляет том вместе с данными.

## API

| Метод и путь | Действие | Успешный ответ |
| --- | --- | --- |
| `POST /auth/register` | Создать аккаунт с полями `email` и `password` | `201 Created` |
| `POST /auth/login` | Войти и получить cookie сессии | `204 No Content` |
| `POST /auth/logout` | Завершить текущую сессию | `204 No Content` |
| `GET /tasks` | Получить список задач, отсортированный по ID | `200 OK`, объект с полем `tasks` |
| `GET /tasks?done=true` | Показать только выполненные задачи; для невыполненных используй `done=false` | `200 OK` |
| `POST /tasks` | Создать задачу с полем `title` | `201 Created`, объект с ID задачи |
| `PATCH /tasks/{id}` | Изменить название или отметить задачу выполненной | `204 No Content` |
| `DELETE /tasks/{id}` | Удалить задачу по ID | `204 No Content` |

### Примеры запросов

Все маршруты `/tasks` требуют cookie действующей сессии. Пример для нового локального аккаунта:

```bash
cookie_file=$(mktemp)
curl -i http://127.0.0.1:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"local-demo-password"}'

curl -i -c "$cookie_file" http://127.0.0.1:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"demo@example.com","password":"local-demo-password"}'
```

Создать задачу:

```bash
curl -i -b "$cookie_file" -X POST http://127.0.0.1:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Изучить Go"}'
```

Пример ответа: `{"id":1}`. В командах ниже замени `1` на полученный ID.

```bash
# Отметить задачу выполненной
curl -i -b "$cookie_file" -X PATCH http://127.0.0.1:8080/tasks/1 \
  -H 'Content-Type: application/json' \
  -d '{"done":true}'

# Показать выполненные задачи
curl -i -b "$cookie_file" 'http://127.0.0.1:8080/tasks?done=true'

# Удалить задачу
curl -i -b "$cookie_file" -X DELETE http://127.0.0.1:8080/tasks/1

# Завершить сессию
curl -i -b "$cookie_file" -X POST http://127.0.0.1:8080/auth/logout
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

При создании БД на пустом томе PostgreSQL берёт пароль из `POSTGRES_PASSWORD` и выполняет [init.sql](init.sql), создавая таблицы `users`, `tasks` и `sessions`. На уже инициализированном томе изменение этих настроек не меняет пароль и не применяет SQL повторно. Если том создан до появления аккаунтов, схему нужно обновить миграцией: добавить таблицы аккаунтов и сессий и назначить владельцев существующим задачам.

## Структура проекта

| Файл или каталог | Назначение |
| --- | --- |
| [main.go](main.go) | Подключение к БД, регистрация маршрутов и запуск HTTP-сервера |
| [handlers/](handlers/) | Обработка HTTP-запросов, проверка входных данных и ответы в JSON |
| [repository/](repository/) | SQL-запросы к PostgreSQL |
| [tasks/](tasks/) | Структура задачи |
| [users/](users/) | Модель данных для входа и регистрации |
| [web/](web/), [frontend.go](frontend.go) | Веб-интерфейс на HTML, CSS и JavaScript, встроенный в Go-сервер |
| [init.sql](init.sql) | Создание таблиц при инициализации базы |
| [Dockerfile](Dockerfile), [docker-compose.yaml](docker-compose.yaml) | Сборка приложения и запуск контейнеров |
