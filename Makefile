# Docker Compose arguments
CONFIG = docker-compose.yml

all: build run

build:
	docker compose -f $(CONFIG) build

run:
	docker compose -f $(CONFIG) up -d --build

stop:
	docker compose -f $(CONFIG) down

clean:
	docker compose -f $(CONFIG) down --rmi all --remove-orphans