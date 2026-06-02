# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Компилируем и сервер, и импортер
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o importer ./cmd/import

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/importer .
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/handlers/templates ./internal/handlers/templates
COPY --from=builder /app/pkg/questions ./data
COPY --from=builder /app/uploads ./uploads

# Добавляем переменные окружения ПРЯМО ЗДЕСЬ
ENV DB_HOST=amvera-egorovandrey-cnpg-medbookdb-rw
ENV DB_PORT=5432
ENV DB_USER=atlas
ENV DB_PASSWORD=123
ENV DB_NAME=med_book

# Создаем entrypoint скрипт
RUN echo '#!/bin/sh' > entrypoint.sh && \
    echo 'echo "📥 Running data import..."' >> entrypoint.sh && \
    echo './importer -dsn "host=$DB_HOST user=$DB_USER password=$DB_PASSWORD dbname=$DB_NAME port=$DB_PORT sslmode=disable" -file ./data/questions.json -skip true' >> entrypoint.sh && \
    echo 'echo "🚀 Starting server..."' >> entrypoint.sh && \
    echo 'exec ./server' >> entrypoint.sh && \
    chmod +x entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
