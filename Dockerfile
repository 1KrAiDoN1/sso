FROM golang:1.24.3-alpine AS builder

RUN apk add --no-cache git build-base

# Устанавливаем migrate
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /sso

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /sso/sso ./cmd/sso/main.go

# Финальный образ
FROM alpine:latest

WORKDIR /sso

# Устанавливаем инструменты для работы с БД
RUN apk add --no-cache postgresql-client bash

# Копируем бинарники, миграции, конфиг и скрипт запуска
COPY --from=builder /sso/sso .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY --from=builder /sso/migrations ./migrations
COPY --from=builder /sso/internal/config /sso/internal/config
COPY ./entrypoint.sh .

# Делаем скрипт исполняемым внутри контейнера
RUN chmod +x ./entrypoint.sh

EXPOSE 4404

# Устанавливаем команду по умолчанию
CMD ["./entrypoint.sh"]