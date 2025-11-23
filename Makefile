.PHONY: help build run test clean docker-build docker-up docker-down docker-logs migrate-up migrate-down

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build:
	@go build -o bin/reviewer-service ./cmd/reviewer-service

run:
	@CONFIG_PATH=./config/local.yaml ./bin/reviewer-service

test:
	@go test -v ./...

test-integration:
	@go test -v ./internal/tests/integration/...

clean:
	@rm -rf bin/
	@go clean

docker-build:
	@docker-compose build

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

docker-logs:
	@docker-compose logs -f

docker-restart:
	@docker-compose restart

migrate-up:
	@docker-compose run --rm migrate -path=/migrations -database="postgres://reviewer:reviewer_password@postgres:5432/reviewer_db?sslmode=disable" up

migrate-down:
	@docker-compose run --rm migrate -path=/migrations -database="postgres://reviewer:reviewer_password@postgres:5432/reviewer_db?sslmode=disable" down

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@docker-compose run --rm migrate create -ext sql -dir /migrations -seq $(NAME)

deps:
	@go mod download
	@go mod tidy

lint:
	@golangci-lint run || echo "golangci-lint not installed, skipping..."

fmt:
	@go fmt ./...

vet:
	@go vet ./...

