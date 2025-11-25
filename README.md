# golang-dining-ordering
heygreet clone - backend for ordering at the restaurant

[![Go CI](https://github.com/simonasbuj/golang-dining-ordering/actions/workflows/ci.yml/badge.svg)](https://github.com/simonasbuj/golang-dining-ordering/actions/workflows/ci.yml)
[![codecov](https://codecov.io/github/simonasbuj/golang-dining-ordering/graph/badge.svg?token=0Z1QP6KJYZ)](https://codecov.io/github/simonasbuj/golang-dining-ordering)

### Dev Dependencies
- **[golang-migrate](https://github.com/golang-migrate/migrate)** – Database migrations
- **[sqlc](https://github.com/sqlc-dev/sqlc)** – Generate models and DB functions from migration files

### Pre-Commit Tools
- **[golangci-lint](https://github.com/golangci/golangci-lint)** – Linter for Go code


## Running the API

You can run the API in different ways:

### Directly with Go
```bash
go run cmd/main.go
```

### Using Docker
```bash
docker compose up -d --build
```

### Make commands
```bash
make run-api
```

```bash
make run-all
```

## Architecture

![alt text](assets/images/architecture-diagram.png)
