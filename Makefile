include .env
export $(shell sed 's/=.*//' .env)
DOCKER_COMPOSE = docker-compose
DB_URL = postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable
migrate-up:
		migrate -path ./migrations -database "${DB_URL}" up
migrate-down:
		yes | migrate -path ./migrations -database "${DB_URL}" down
docker-up:
		$(DOCKER_COMPOSE) up -d

docker-build:
		$(DOCKER_COMPOSE) build

docker-down:
		$(DOCKER_COMPOSE) down
docker-logs:
		$(DOCKER_COMPOSE) logs -f
