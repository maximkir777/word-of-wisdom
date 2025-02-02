# Makefile

.PHONY: help run-server run-client docker-up docker-down env test lint format clean build-server build-client install-bin-deps

help:
	@echo "Usage: make <target>"
	@echo "  run-server         - Run the server locally"
	@echo "  run-client         - Run the client locally"
	@echo "  docker-up          - Build and start Docker containers"
	@echo "  docker-down        - Stop Docker containers"
	@echo "  env                - Generate .env file from .env.example if not exists"
	@echo "  test               - Run tests"
	@echo "  lint               - Run golangci-lint"
	@echo "  format             - Run go fmt"
	@echo "  build-server       - Build server binary into bin directory"
	@echo "  build-client       - Build client binary into bin directory"
	@echo "  install-bin-deps   - Install binary dependencies (linters, etc.)"
	@echo "  clean              - Remove built binaries"

run-server:
	@echo "Running server locally..."
	go run ./cmd/server/main.go

run-client:
	@echo "Running client locally..."
	go run ./cmd/client/main.go

docker-up:
	@echo "Starting Docker containers..."
	docker-compose up --build

docker-down:
	@echo "Stopping Docker containers..."
	docker-compose down

env:
	@test -f .env || cp .env.example .env

test:
	@echo "Running tests..."
	go test ./... -v

lint:
	@echo "Running golangci-lint..."
	golangci-lint run

format:
	@echo "Running go fmt..."
	go fmt ./...
	@echo "Running smartimports..."
	smartimports -local .

build-server:
	@echo "Building server binary..."
	mkdir -p bin
	go build -o bin/server ./cmd/server/main.go

build-client:
	@echo "Building client binary..."
	mkdir -p bin
	go build -o bin/client ./cmd/client/main.go

install-bin-deps:
	@echo "Installing binary dependencies..."
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing smartimports..."
	go install github.com/pav5000/smartimports/cmd/smartimports@v0.2.0

clean:
	@echo "Cleaning built binaries..."
	rm -rf bin
