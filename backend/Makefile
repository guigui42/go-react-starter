.PHONY: lint fmt test test-ci test-coverage run dev build vet

# Default target
.DEFAULT_GOAL := run

lint: ## Run Go linters
golangci-lint run

fmt: ## Format Go code
go fmt ./...

test: ## Run all tests
JWT_SECRET=test-secret-for-testing-minimum-32-chars ENCRYPTION_KEY=test-encryption-key-32-chars go test -v -race $(shell go list ./... | grep -v /tools)

test-ci: ## Run tests in CI mode (short, parallel, with coverage)
JWT_SECRET=test-secret-for-testing-minimum-32-chars ENCRYPTION_KEY=test-encryption-key-32-chars go test -short -race -p 8 -parallel 8 -coverprofile=coverage.out -covermode=atomic $(shell go list ./... | grep -v /tools)

test-coverage: ## Run tests with coverage report
JWT_SECRET=test-secret-for-testing-minimum-32-chars ENCRYPTION_KEY=test-encryption-key-32-chars go test -v -race -coverprofile=coverage.out -covermode=atomic $(shell go list ./... | grep -v /tools)
go tool cover -html=coverage.out -o coverage.html
@echo "Coverage report generated: coverage.html"
@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

run: ## Run the API server
go run cmd/server/main.go

dev: run ## Alias for run

build: ## Build the server binary
go build -o server ./cmd/server

vet: ## Run go vet
go vet ./...
