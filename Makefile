.PHONY: help dev dev-observability dev-backend dev-frontend install setup update update-backend update-frontend test test-backend test-frontend test-e2e lint lint-backend lint-frontend fmt clean release db-up db-down db-restart db-logs db-shell db-reset db-init db-wait

.DEFAULT_GOAL := help

BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m

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
	@echo "  make update           - Update all dependencies"
	@echo ""
	@echo "$(BLUE)Testing:$(NC)"
	@echo "  make test             - Run all tests (backend + frontend)"
	@echo "  make test-backend     - Run backend tests with coverage"
	@echo "  make test-frontend    - Run frontend tests with coverage"
	@echo "  make test-e2e         - Run end-to-end tests (Playwright)"
	@echo ""
	@echo "$(BLUE)Code Quality:$(NC)"
	@echo "  make lint             - Run all linters"
	@echo "  make fmt              - Format all code"
	@echo ""
	@echo "$(BLUE)Database:$(NC)"
	@echo "  make db-up            - Start local PostgreSQL container"
	@echo "  make db-down          - Stop local PostgreSQL container"
	@echo "  make db-restart       - Restart PostgreSQL container"
	@echo "  make db-logs          - Tail PostgreSQL logs"
	@echo "  make db-shell         - Open psql shell"
	@echo "  make db-reset         - Recreate PostgreSQL volume (destructive)"
	@echo "  make db-init          - Start DB and wait for readiness"
	@echo ""
	@echo "$(BLUE)Other:$(NC)"
	@echo "  make release          - Trigger release workflow"
	@echo "  make clean            - Remove build artifacts"
	@echo ""

dev:
	@echo "$(GREEN)Starting development environment...$(NC)"
	@echo "$(YELLOW)Backend:  http://localhost:8080$(NC)"
	@echo "$(YELLOW)Frontend: http://localhost:5173$(NC)"
	@echo ""
	@trap 'kill 0' EXIT; \
		$(MAKE) dev-backend & \
		$(MAKE) dev-frontend & \
		wait

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

dev-backend:
	@echo "$(BLUE)Starting backend server...$(NC)"
	@cd backend && $(MAKE) run

dev-frontend:
	@echo "$(BLUE)Starting frontend dev server...$(NC)"
	@cd frontend && $(MAKE) dev

install:
	@echo "$(GREEN)Installing dependencies...$(NC)"
	@echo "$(BLUE)Backend...$(NC)"
	@cd backend && go mod download
	@echo "$(BLUE)Frontend...$(NC)"
	@cd frontend && npm install
	@echo "$(GREEN)✓ All dependencies installed$(NC)"

setup: install db-up db-wait
	@echo ""
	@echo "$(GREEN)✓ Setup complete!$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. cp backend/.env.example backend/.env"
	@echo "  2. Edit backend/.env (set JWT_SECRET and ENCRYPTION_KEY)"
	@echo "  3. make dev"
	@echo ""

update: update-backend update-frontend
	@echo "$(GREEN)✓ All dependencies updated$(NC)"

update-backend:
	@echo "$(BLUE)Updating backend dependencies...$(NC)"
	@cd backend && go get -u ./...
	@cd backend && go mod tidy
	@echo "$(GREEN)✓ Backend dependencies updated$(NC)"

update-frontend:
	@echo "$(BLUE)Updating frontend dependencies...$(NC)"
	@command -v ncu >/dev/null 2>&1 || { echo "$(YELLOW)Installing npm-check-updates...$(NC)"; npm install -g npm-check-updates; }
	@cd frontend && ncu --target minor -u
	@cd frontend && npm install
	@echo "$(GREEN)✓ Frontend dependencies updated$(NC)"

test: test-backend test-frontend
	@echo "$(GREEN)✓ All tests passed$(NC)"

test-backend:
	@echo "$(BLUE)Running backend tests...$(NC)"
	@cd backend && $(MAKE) test-coverage

test-frontend:
	@echo "$(BLUE)Running frontend tests...$(NC)"
	@cd frontend && npm run test:coverage

test-e2e:
	@echo "$(BLUE)Running end-to-end tests...$(NC)"
	@cd frontend && npm run test:e2e

lint: lint-backend lint-frontend
	@echo "$(GREEN)✓ All linting checks passed$(NC)"

lint-backend:
	@echo "$(BLUE)Running backend linters...$(NC)"
	@cd backend && $(MAKE) lint

lint-frontend:
	@echo "$(BLUE)Running frontend linters...$(NC)"
	@cd frontend && npm run lint

fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	@cd backend && $(MAKE) fmt
	@cd frontend && npm run format
	@echo "$(GREEN)✓ Code formatted$(NC)"

db-up:
	@echo "$(BLUE)Starting local PostgreSQL (port $${DB_PORT:-5433})...$(NC)"
	@docker compose -f docker/docker-compose.yml up -d db 2>&1 || true
	@echo "$(GREEN)✓ PostgreSQL started$(NC)"

db-down:
	@echo "$(BLUE)Stopping local PostgreSQL...$(NC)"
	@docker compose -f docker/docker-compose.yml stop db
	@echo "$(GREEN)✓ PostgreSQL stopped$(NC)"

db-restart: db-down db-up
	@echo "$(GREEN)✓ PostgreSQL restarted$(NC)"

db-logs:
	@docker compose -f docker/docker-compose.yml logs -f db

db-shell:
	@docker compose -f docker/docker-compose.yml exec db psql -U postgres -d starter

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

db-init: db-up db-wait
	@echo "$(GREEN)✓ Local database initialized$(NC)"

db-reset:
	@echo "$(YELLOW)Resetting local PostgreSQL (this deletes all data)...$(NC)"
	@docker compose -f docker/docker-compose.yml down -v
	@docker compose -f docker/docker-compose.yml up -d db
	@$(MAKE) db-wait
	@echo "$(GREEN)✓ PostgreSQL reset complete$(NC)"

NOTES ?=
release:
	@echo "$(GREEN)Creating new release...$(NC)"
	@if [ -z "$(NOTES)" ]; then \
		gh workflow run release.yml; \
	else \
		gh workflow run release.yml -f release_notes="$(NOTES)"; \
	fi
	@echo "$(GREEN)✓ Release workflow triggered$(NC)"

clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@cd backend && rm -f server coverage.out coverage.html
	@cd frontend && rm -rf node_modules dist coverage .vite
	@echo "$(GREEN)✓ Clean complete$(NC)"
