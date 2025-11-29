FROM golang:1.25.3-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o dineApp ./cmd/api 

RUN chmod +x /app/dineApp

# build tiny docker image
FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/dineApp /app
COPY --from=builder /app/api/openapi-spec ./api/openapi-spec
COPY --from=builder /app/frontend ./frontend

CMD ["/app/dineApp"]
