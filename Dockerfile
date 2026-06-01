# syntax=docker/dockerfile:1

# ---- 1. Стадия сборки (builder) ----
# Здесь мы компилируем наше Go-приложение
FROM golang:1.23-alpine AS builder

# Указываем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы с зависимостями
COPY go.mod go.sum ./
# Скачиваем их
RUN go mod download

# Копируем весь остальной код проекта
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api

# ---- 2. Финальная стадия (runner) ----
# Создаем минимальный образ, в котором будет только наш скомпилированный бинарник
FROM alpine:latest

# Устанавливаем CA-сертификаты (нужны, если твой сайт ходит по HTTPS к другим сервисам)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Копируем скомпилированный бинарник из стадии "builder"
COPY --from=builder /app/server .

COPY --from=builder /app/internal/handlers/templates ./internal/handlers/templates

COPY --from=builder /app/static ./static

COPY --from=builder /app/uploads ./uploads

COPY --from=builder /app/pkg ./pkg

# Сообщаем Docker, что приложение будет слушать порт 8080 (или тот, который используешь ты)
EXPOSE 8080

# Команда для запуска сервера
CMD ["./server"]