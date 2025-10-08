# golang-dining-ordering
heygreet clone - backend for ordering at the restaurant


## Running API

### simple
```
go run cmd/main.go
```
```
make run-api
```

### run with listening to changes using air
Istall air with:
```
go install github.com/air-verse/air@latest
export PATH=$PATH:$(go env GOPATH)/bin

```

Run:
```
air
```
```
make run-api-with-ari
```

## Dev dependencies
`golang-migrations` used for database migrations:
```
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

`sqlc` used to generate models and db insert/update/delete functions using migration files:
```
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

## Pre-Commit
### golangci-lint
Install:
```
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

Run:
```
golangci-lint run --verbose --fix

make lint
```
