MIGRATIONS_DIR := ./services/management/db/sql/migrations
DINE_DB_URI := potsgres-uri

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
	go run ./cmd/api

run-api-with-air:
	air

run-all:
	docker compose up -d --build	

# pre-commit
lint:
	golangci-lint run --verbose --max-issues-per-linter=0 --max-same-issues=0

lint-fix:
	golangci-lint run --verbose --fix

.PHONY: test
test:
	go test -v -coverprofile=cvr.txt ./... && grep -v -e "/generated/" -e "/repository/" -e "/mock/" -e "/cmd/" -e "/routes/" cvr.txt > coverage.txt

.PHONY: test-race
test-race:
	go test -v -race -coverprofile=cvr.txt ./... && grep -v -e "/generated/" -e "/repository/" -e "/mock/" -e "/cmd/" -e "/routes/" cvr.txt > coverage.txt

cov-html:
	go tool cover -html=coverage.txt -o coverage.html

.PHONY: coverage
coverage:
	go tool cover -func=coverage.txt



# database
start-postgres:
	docker-compose -f infra/docker/postgres-docker-compose.yml up -d

stop-postgres:
	docker-compose -f infra/docker/postgres-docker-compose.yml stop -d

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

up-migrations:
	echo "$(DINE_DB_URI)"
	migrate -path $(MIGRATIONS_DIR) -database "$($(DINE_DB_URI_NAME))" up

down-migrations:
	migrate -path $(MIGRATIONS_DIR) -database "$($(DINE_DB_URI_NAME))" down 1

sqlc-generate:
	sqlc generate

up-all-migrations:
	@echo "Migrate auth"
	migrate -path "./services/auth/db/sql/migrations" -database "$(DINE_AUTH_DB_URI)" up

	@echo "Migrate management"
	migrate -path "./services/management/db/sql/migrations" -database "$(DINE_MANAGEMENT_DB_URI)" up
	
	@echo "Migrate orders"
	migrate -path "./services/orders/db/sql/migrations" -database "$(DINE_ORDERS_DB_URI)" up


down-all-migrations:
	@echo "Migrate down auth"
	migrate -path "./services/orders/db/sql/migrations" -database "$(DINE_ORDERS_DB_URI)" down

	@echo "Migrate down management"
	migrate -path "./services/management/db/sql/migrations" -database "$(DINE_MANAGEMENT_DB_URI)" down

	@echo "Migrate down auth"
	migrate -path "./services/auth/db/sql/migrations" -database "$(DINE_AUTH_DB_URI)" down

# s3 storage
start-minio:
	docker-compose -f infra/docker/minio-docker-compose.yml up -d

stop-minio:
	docker-compose -f infra/docker/minio-docker-compose.yml stop
