.PHONY: dev build preview test test-coverage lint format install type-check

# Default target
.DEFAULT_GOAL := dev

install: ## Install npm dependencies
npm install

dev: ## Run development server
npm run dev

build: ## Build for production
npm run build

preview: ## Preview production build
npm run preview

test: ## Run unit tests
npm run test

test-coverage: ## Run tests with coverage
npm run test:coverage

lint: ## Run ESLint
npm run lint

format: ## Format code with Prettier
npm run format

type-check: ## Run TypeScript type checking
npx tsc --noEmit
