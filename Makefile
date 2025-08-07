include .env
export $(shell sed 's/=.*//' .env)
DB_URL = postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
migrate-up:
		migrate -path ./migrations -database "${DB_URL}" up
migrate-down:
		yes | migrate -path ./migrations -database "${DB_URL}" down
