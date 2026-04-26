.PHONY: dev dev-backend dev-frontend install test lint build

# Development
dev: ## Run both backend and frontend in dev mode
@echo "Starting development servers..."
@make -j2 dev-backend dev-frontend

dev-backend: ## Run backend dev server
cd backend && make run

dev-frontend: ## Run frontend dev server
cd frontend && npm run dev

# Dependencies
install: ## Install all dependencies
cd backend && go mod download
cd frontend && npm install

# Testing
test: ## Run all tests
cd backend && make test
cd frontend && npm run test -- --run

# Linting
lint: ## Run all linters
cd backend && make lint
cd frontend && npm run lint

# Building
build: ## Build both backend and frontend
cd backend && make build
cd frontend && npm run build

help: ## Show this help
@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
