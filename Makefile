.PHONY: up down build rebuild logs migrate test run tidy psql

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build --no-cache

rebuild: down build up

logs:
	docker compose logs -f bot

migrate:
	docker compose exec -T postgres psql -U postgres payment_bot < migrations/001_init.sql

test:
	go test ./...

run:
	go run ./cmd/bot/

tidy:
	go mod tidy

psql:
	docker compose exec postgres psql -U postgres payment_bot
