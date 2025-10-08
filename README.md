


# Running API

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

# Pre-Commit
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
