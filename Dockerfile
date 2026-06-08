# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -o importer ./cmd/import

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/server .
COPY --from=builder /app/importer .
COPY --from=builder /app/static ./static
COPY --from=builder /app/internal/handlers/templates ./internal/handlers/templates
COPY --from=builder /app/pkg/questions ./pkg/questions
COPY --from=builder /app/uploads ./uploads

ENV DB_DRIVER=sqlite
ENV DB_DSN=/data/med_book.db?_journal_mode=WAL&_foreign_keys=1
ENV JWT_SECRET=change-me-in-production
ENV SERVER_PORT=8080

RUN mkdir -p /data && \
    echo '#!/bin/sh' > entrypoint.sh && \
    echo 'echo "Running data import..."' >> entrypoint.sh && \
    echo './importer -dsn "/data/med_book.db?_journal_mode=WAL&_foreign_keys=1" -file ./pkg/questions/questions.json -skip true' >> entrypoint.sh && \
    echo 'echo "Starting server..."' >> entrypoint.sh && \
    echo 'exec ./server' >> entrypoint.sh && \
    chmod +x entrypoint.sh

EXPOSE 8080

VOLUME ["/data"]

ENTRYPOINT ["./entrypoint.sh"]
