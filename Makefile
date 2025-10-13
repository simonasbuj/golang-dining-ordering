MIGRATIONS_DIR := ./internal/db/sql/migrations

# make sure golang dependencies likq 'sqlc' and 'migrations' are in the path
PATH := $(PATH):$(shell go env GOPATH)/bin
export PATH

# Load .env.example and .env
ifneq (,$(wildcard .env.example))
    include .env.example
    export $(shell sed 's/=.*//' .env.example)
endif

ifneq (,$(wildcard .env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif


# run app
run-api:
	go run cmd/main.go

run-api-with-air:
	air

# pre-commit
lint:
	golangci-lint run --verbose

lint-fix:
	golangci-lint run --verbose --fix

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
	echo "$(DB_URI)"
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URI)" up

down-migrations:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URI)" down 1

sqlc-generate:
	sqlc generate
