run-api:
	go run cmd/main.go

run-api-with-air:
	air

lint:
	golangci-lint run --verbose --fix
