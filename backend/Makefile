.PHONY: lint fmt test test-ci test-coverage run dev build vet

.DEFAULT_GOAL := run

lint:
	golangci-lint run

fmt:
	go fmt ./...

test:
	JWT_SECRET=test-secret-for-testing-minimum-32-chars ENCRYPTION_KEY=test-encryption-key-32-chars go test -v -race $$(go list ./... | grep -v /tools)

test-ci:
	JWT_SECRET=test-secret-for-testing-minimum-32-chars ENCRYPTION_KEY=test-encryption-key-32-chars go test -short -race -p 8 -parallel 8 -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v /tools)

test-coverage:
	JWT_SECRET=test-secret-for-testing-minimum-32-chars ENCRYPTION_KEY=test-encryption-key-32-chars go test -v -race -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v /tools)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

run:
	go run cmd/server/main.go

dev: run

build:
	go build -o server ./cmd/server

vet:
	go vet ./...
