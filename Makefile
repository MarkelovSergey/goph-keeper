BINARY_SERVER=goph-keeper-server
BINARY_CLIENT=goph-keeper

# Подгружаем переменные из .env файла, если он существует
-include .env
export

MODULE=github.com/MarkelovSergey/goph-keeper
VERSION?=0.0.1
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LD_FLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildDate=$(BUILD_DATE)"

.PHONY: all build build-server build-client test lint migrate-up migrate-down swag run-server run-client docker-up docker-down clean

all: build

## Build
build: build-server build-client

build-server:
	go build $(LD_FLAGS) -o bin/$(BINARY_SERVER) ./cmd/server/

build-client:
	go build $(LD_FLAGS) -o bin/$(BINARY_CLIENT) ./cmd/client/

## Cross-platform builds
build-all:
	GOOS=linux   GOARCH=amd64 go build $(LD_FLAGS) -o bin/$(BINARY_CLIENT)-linux-amd64   ./cmd/client/
	GOOS=darwin  GOARCH=amd64 go build $(LD_FLAGS) -o bin/$(BINARY_CLIENT)-darwin-amd64  ./cmd/client/
	GOOS=windows GOARCH=amd64 go build $(LD_FLAGS) -o bin/$(BINARY_CLIENT)-windows-amd64.exe ./cmd/client/

## Test
test:
	go test ./... -count=1

test-cover:
	go test ./... -count=1 -coverprofile=coverage.out
	go tool cover -func=coverage.out

## Lint
lint:
	golangci-lint run ./...

## Migrations (DATABASE_DSN must be set)
.PHONY: install-migrate
install-migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

migrate-up:
	$(shell go env GOPATH)/bin/migrate -path migrations -database "$(DATABASE_DSN)" up

migrate-down:
	$(shell go env GOPATH)/bin/migrate -path migrations -database "$(DATABASE_DSN)" down

## Swagger
swag:
	swag init -g cmd/server/main.go -o docs/

## Run
run-server:
	go run ./cmd/server/

run-client:
	go run ./cmd/client/

## Docker
docker-up:
	docker compose up -d

docker-down:
	docker compose down

## Clean
clean:
	rm -rf bin/ coverage.out
