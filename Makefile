SHELL := /bin/sh

MIGRATE_DB_URL ?= postgres://moneymate:Da789852!..@localhost:5433/expense_db?sslmode=disable

.PHONY: dev-up dev-down prod-up prod-down sqlc migrate-up migrate-down migrate-version migrate-create

db-up:
	docker compose up -d db redis

dev-up:
	docker compose up -d

dev-down:
	docker compose down

prod-up:
	docker compose -f docker-compose.yaml -f docker-compose.prod.yml up -d

prod-down:
	docker compose -f docker-compose.yaml -f docker-compose.prod.yml down

sqlc:
	sqlc generate

migrate-up:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" down 1

migrate-version:
	migrate -path migrations -database "$(MIGRATE_DB_URL)" version

migrate-create:
	@read -p "name: " name; migrate create -ext sql -dir migrations -seq $$name