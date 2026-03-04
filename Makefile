SHELL := /bin/sh

MIGRATE_DB_URL ?= postgres://moneymate:Da789852!..@localhost:5433/expense_db?sslmode=disable

.PHONY: dev-up dev-down prod-up prod-down sqlc migrate-up migrate-down migrate-version migrate-create test test-integration

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

test:
	MIGRATE_DB_URL="$(MIGRATE_DB_URL)" go test ./... -v -cover

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

MOCKGEN := mockgen

mocks:
	@set -e; \
	for src in $$(find internal -type f -path "*/interfaces/*.go"); do \
		dir=$$(dirname "$$src"); \
		base=$$(basename "$$src" .go); \
		out_dir="$$dir/mocks"; \
		out_file="$$out_dir/$${base}_mock.go"; \
		mkdir -p "$$out_dir"; \
		echo "Generating $$out_file from $$src"; \
		$(MOCKGEN) -source="$$src" -destination="$$out_file" -package="mocks"; \
	done

mocks-clean:
	@find internal -type f -path "*/interfaces/mocks/*_mock.go" -delete