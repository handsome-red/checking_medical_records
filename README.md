# Сервис тестирования медицинских книжек

Веб-приложение для сотрудников Роспотребнадзора: проверка знаний по корректности оформления медицинских книжек.

## Стек

- Go (chi, GORM)
- SQLite
- HTML + JavaScript (server-side templates)

## Запуск локально

```bash
go run ./cmd/import/ -file pkg/questions/questions.json -skip
go run ./cmd/api/
```

Сервер: http://localhost:8080

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `DB_DRIVER` | `sqlite` | Драйвер БД |
| `DB_DSN` | `./data/med_book.db?...` | Путь к SQLite |
| `JWT_SECRET` | `change-me-in-production` | Секрет JWT |
| `ADMIN_EMAIL` | — | Email пользователя-администратора |
| `SERVER_PORT` | `8080` | Порт HTTP |

## Администратор

Укажите `ADMIN_EMAIL` при регистрации или входе — пользователь получит роль admin и доступ к `/admin` (просмотр результатов, фильтры, экспорт Excel).

## Docker

```bash
docker compose up --build
```
