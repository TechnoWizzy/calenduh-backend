# Docker Compose arguments
CONFIG = docker-compose.yml
POSTGRES_DIR = postgres/migrations

all: build run

build:
	docker compose -f $(CONFIG) build

run:
	docker compose -f $(CONFIG) up -d --build

stop:
	docker compose -f $(CONFIG) down

clean:
	docker compose -f $(CONFIG) down --rmi all --remove-orphans

sqlc:
	sqlc generate

migration:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(POSTGRES_DIR) $$name

clean-db:
	docker compose -f $(CONFIG) down --volumes