#!/bin/sh

# Выходим из скрипта, если любая команда завершилась с ошибкой
set -e

# Формируем URL для подключения к БД
# Обратите внимание, что мы используем хост 'db', как имя сервиса в docker-compose
DB_URL="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@db:5432/${POSTGRES_DB}?sslmode=disable"

echo "Waiting for database to be ready..."
# Ждем, пока база данных не станет доступной
# pg_isready ждет, пока PostgreSQL не начнет принимать соединения
while ! pg_isready -h db -p 5432 -U "${POSTGRES_USER}"; do
  sleep 2
done
echo "Database is ready!"

echo "Applying migrations..."
# Применяем миграции
# -path указывает на директорию с файлами миграций
# -database указывает на URL для подключения к БД
migrate -path /sso/migrations -database "${DB_URL}" up

echo "Migrations applied successfully!"

echo "Starting application..."
# Запускаем основное приложение
# exec заменяет текущий процесс (sh) на процесс приложения, что правильно для Docker
exec ./sso