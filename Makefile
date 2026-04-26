.PHONY: help dev dev-observability dev-backend dev-frontend install setup update update-backend update-frontend test test-backend test-frontend test-e2e lint lint-backend lint-frontend fmt clean release db-up db-down db-restart db-logs db-shell db-reset db-init

# Default target
.DEFAULT_GOAL := help

# Colors for terminal output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

## help: Display this help message
help:
@echo "$(GREEN)Go React Starter - Development Commands$(NC)"
@echo ""
@echo "$(BLUE)Development:$(NC)"
@echo "  make dev                - Run both backend and frontend in parallel"
@echo "  make dev-observability  - Run dev with metrics + tracing enabled"
@echo "  make dev-backend        - Run only the backend server (port 8080)"
@echo "  make dev-frontend       - Run only the frontend dev server (port 5173)"
@echo ""
@echo "$(BLUE)Setup:$(NC)"
@echo "  make install          - Install all dependencies (backend + frontend)"
@echo "  make setup            - Full project setup (install + DB + migrations)"
@echo "  make update           - Update all dependencies (backend + frontend)"
@echo "  make update-backend   - Update Go dependencies"
@echo "  make update-frontend  - Update npm dependencies (requires ncu)"
@echo ""
@echo "$(BLUE)Testing:$(NC)"
@echo "  make test             - Run all tests (backend + frontend)"
@echo "  make test-backend     - Run backend tests with coverage"
@echo "  make test-frontend    - Run frontend tests with coverage"
@echo "  make test-e2e         - Run end-to-end tests (Playwright)"
@echo ""
@echo "$(BLUE)Code Quality:$(NC)"
@echo "  make lint             - Run linters for both backend and frontend"
@echo "  make lint-backend     - Run Go linters"
@echo "  make lint-frontend    - Run TypeScript/React linters"
@echo "  make fmt              - Format all code"
@echo ""
@echo "$(BLUE)Database:$(NC)"
@echo "  make db-up            - Start local PostgreSQL container"
@echo "  make db-down          - Stop local PostgreSQL container"
@echo "  make db-restart       - Restart local PostgreSQL container"
@echo "  make db-logs          - Tail PostgreSQL container logs"
@echo "  make db-shell         - Open psql shell in PostgreSQL container"
@echo "  make db-reset         - Recreate local PostgreSQL volume (destructive)"
@echo "  make db-init          - Start DB and wait for readiness"
@echo ""
@echo "$(BLUE)Observability:$(NC)"
@echo "  make dev-observability - Run dev with OTel metrics + tracing + logs"
@echo ""
@echo "$(BLUE)Release:$(NC)"
@echo "  make release          - Trigger release workflow via GitHub CLI"
@echo ""
@echo "$(BLUE)Cleanup:$(NC)"
@echo "  make clean            - Remove build artifacts and dependencies"
@echo ""

# ──────────────────────────────────────────────────────────────
# Development
# ──────────────────────────────────────────────────────────────

## dev: Run both backend and frontend in parallel
dev:
@echo "$(GREEN)Starting development environment...$(NC)"
@echo "$(YELLOW)Backend:  http://localhost:8080$(NC)"
@echo "$(YELLOW)Frontend: http://localhost:5173$(NC)"
@echo ""
@trap 'kill 0' EXIT; \
$(MAKE) dev-backend & \
$(MAKE) dev-frontend & \
wait

## dev-observability: Run dev with observability enabled (metrics + tracing + logs)
dev-observability:
@echo "$(GREEN)Starting with observability enabled...$(NC)"
@echo "$(YELLOW)Backend:  http://localhost:8080$(NC)"
@echo "$(YELLOW)Frontend: http://localhost:5173$(NC)"
@echo "$(YELLOW)Metrics:  http://localhost:8080/metrics$(NC)"
@echo ""
@trap 'kill 0' EXIT; \
OTEL_ENABLED=true PROMETHEUS_ENABLED=true OTEL_TRACING_ENABLED=true OTEL_LOGS_ENABLED=true OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 OTEL_EXPORTER_OTLP_INSECURE=true $(MAKE) dev-backend & \
$(MAKE) dev-frontend & \
wait

## dev-backend: Run backend server
dev-backend:
@echo "$(BLUE)Starting backend server...$(NC)"
@cd backend && $(MAKE) run

## dev-frontend: Run frontend dev server
dev-frontend:
@echo "$(BLUE)Starting frontend dev server...$(NC)"
@cd frontend && $(MAKE) dev

# ──────────────────────────────────────────────────────────────
# Setup & Dependencies
# ──────────────────────────────────────────────────────────────

## install: Install all dependencies
install:
@echo "$(GREEN)Installing dependencies...$(NC)"
@echo "$(BLUE)Installing backend dependencies...$(NC)"
@cd backend && go mod download
@echo "$(BLUE)Installing frontend dependencies...$(NC)"
@cd frontend && npm install
@echo "$(GREEN)✓ All dependencies installed$(NC)"

## setup: Full project setup (install deps, start DB, wait for readiness)
setup: install db-up db-wait
@echo ""
@echo "$(GREEN)✓ Setup complete!$(NC)"
@echo ""
@echo "$(BLUE)Next steps:$(NC)"
@echo "  1. Copy and configure environment:"
@echo "     cp backend/.env.example backend/.env"
@echo "     (edit backend/.env with your JWT_SECRET and ENCRYPTION_KEY)"
@echo ""
@echo "  2. Start development:"
@echo "     make dev"
@echo ""

## update: Update all dependencies
update: update-backend update-frontend
@echo "$(GREEN)✓ All dependencies updated$(NC)"

## update-backend: Update Go dependencies
update-backend:
@echo "$(BLUE)Updating backend dependencies...$(NC)"
@cd backend && go get -u ./...
@cd backend && go mod tidy
@echo "$(GREEN)✓ Backend dependencies updated$(NC)"

## update-frontend: Update npm dependencies (requires ncu)
update-frontend:
@echo "$(BLUE)Updating frontend dependencies...$(NC)"
@command -v ncu >/dev/null 2>&1 || { echo "$(YELLOW)Installing npm-check-updates globally...$(NC)"; npm install -g npm-check-updates; }
@cd frontend && ncu --target minor -u
@cd frontend && npm install
@echo "$(GREEN)✓ Frontend dependencies updated$(NC)"

# ──────────────────────────────────────────────────────────────
# Testing
# ──────────────────────────────────────────────────────────────

## test: Run all tests
test: test-backend test-frontend
@echo "$(GREEN)✓ All tests passed$(NC)"

## test-backend: Run backend tests
test-backend:
@echo "$(BLUE)Running backend tests...$(NC)"
@cd backend && $(MAKE) test-coverage

## test-frontend: Run frontend tests
test-frontend:
@echo "$(BLUE)Running frontend tests...$(NC)"
@cd frontend && npm run test:coverage

## test-e2e: Run end-to-end tests
test-e2e:
@echo "$(BLUE)Running end-to-end tests...$(NC)"
@cd frontend && npm run test:e2e

# ──────────────────────────────────────────────────────────────
# Code Quality
# ──────────────────────────────────────────────────────────────

## lint: Run all linters
lint: lint-backend lint-frontend
@echo "$(GREEN)✓ All linting checks passed$(NC)"

## lint-backend: Run backend linters
lint-backend:
@echo "$(BLUE)Running backend linters...$(NC)"
@cd backend && $(MAKE) lint

## lint-frontend: Run frontend linters
lint-frontend:
@echo "$(BLUE)Running frontend linters...$(NC)"
@cd frontend && npm run lint

## fmt: Format all code
fmt:
@echo "$(BLUE)Formatting code...$(NC)"
@cd backend && $(MAKE) fmt
@cd frontend && npm run format
@echo "$(GREEN)✓ Code formatted$(NC)"

# ──────────────────────────────────────────────────────────────
# Database
# ──────────────────────────────────────────────────────────────

## db-up: Start local PostgreSQL via Docker Compose
db-up:
@echo "$(BLUE)Starting local PostgreSQL...$(NC)"
@docker compose -f docker/docker-compose.yml up -d db
@echo "$(GREEN)✓ PostgreSQL started$(NC)"

## db-down: Stop local PostgreSQL via Docker Compose
db-down:
@echo "$(BLUE)Stopping local PostgreSQL...$(NC)"
@docker compose -f docker/docker-compose.yml stop db
@echo "$(GREEN)✓ PostgreSQL stopped$(NC)"

## db-restart: Restart local PostgreSQL container
db-restart: db-down db-up
@echo "$(GREEN)✓ PostgreSQL restarted$(NC)"

## db-logs: Tail PostgreSQL container logs
db-logs:
@docker compose -f docker/docker-compose.yml logs -f db

## db-shell: Open psql shell in PostgreSQL container
db-shell:
@docker compose -f docker/docker-compose.yml exec db psql -U postgres -d starter

## db-wait: Wait until local PostgreSQL is ready
db-wait:
@echo "$(BLUE)Waiting for PostgreSQL to be ready...$(NC)"
@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
if docker compose -f docker/docker-compose.yml exec -T db pg_isready -U postgres > /dev/null 2>&1; then \
echo "$(GREEN)✓ PostgreSQL is ready$(NC)"; \
exit 0; \
fi; \
sleep 1; \
done; \
echo "$(YELLOW)⚠ PostgreSQL did not become ready in time$(NC)"; exit 1

## db-init: Start DB and wait for readiness
db-init: db-up db-wait
@echo "$(GREEN)✓ Local database initialized$(NC)"

## db-reset: Recreate local PostgreSQL volume (DESTRUCTIVE)
db-reset:
@echo "$(YELLOW)Resetting local PostgreSQL (this deletes all data)...$(NC)"
@docker compose -f docker/docker-compose.yml down -v
@docker compose -f docker/docker-compose.yml up -d db
@$(MAKE) db-wait
@echo "$(GREEN)✓ PostgreSQL reset complete$(NC)"

# ──────────────────────────────────────────────────────────────
# Release
# ──────────────────────────────────────────────────────────────

## release: Create a new release (triggers Docker build)
NOTES ?=
release:
@echo "$(GREEN)Creating new release...$(NC)"
@if [ -z "$(NOTES)" ]; then \
gh workflow run release.yml; \
else \
gh workflow run release.yml -f release_notes="$(NOTES)"; \
fi
@echo "$(GREEN)✓ Release workflow triggered$(NC)"

# ──────────────────────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────────────────────

## clean: Remove build artifacts and dependencies
clean:
@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
@cd backend && rm -f server coverage.out coverage.html
@cd frontend && rm -rf node_modules dist coverage .vite
@echo "$(GREEN)✓ Clean complete$(NC)"
