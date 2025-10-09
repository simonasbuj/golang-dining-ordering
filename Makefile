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

.PHONY: test
test:
	go test -v ./...

# database
start-postgres:
	docker-compose -f infra/docker/postgres-docker-compose.yml up -d

stop-postgres:
	docker-compose -f infra/docker/postgres-docker-compose.yml stop -d

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

up-migrations:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URI)" up

down-migrations:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URI)" down 1

sqlc-generate:
	sqlc generate
