.PHONY: run test lint build docs migrate

run: ## Run the API server
go run ./cmd/server

test: ## Run all tests
go test ./... -v -count=1

lint: ## Run linter
golangci-lint run

build: ## Build the server binary
go build -o server ./cmd/server

migrate: ## Run database migrations
go run ./cmd/server

vet: ## Run go vet
go vet ./...

.DEFAULT_GOAL := run
