MIGRATIONS_DIR := ./internal/db/sql/migrations

# run app
run-api:
	go run cmd/main.go

run-api-with-air:
	air

# pre-commit
lint-fix:
	golangci-lint run --verbose --fix

lint:
	golangci-lint run --verbose

# database
start-postgres:
	docker-compose -f infra/docker/postgres-docker-compose.yml up -d

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

up-migrations:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URI)" up

down-migration:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URI)" down 1

sqlc-generate:
	sqlc generate
