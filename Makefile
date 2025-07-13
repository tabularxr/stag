.PHONY: build test test-unit test-integration docker-build docker-up docker-down lint fmt clean

APP_NAME := stag
BINARY_NAME := stag
VERSION := 1.0.0

# Build
build:
	go build -o $(BINARY_NAME) ./cmd/$(APP_NAME)

# Testing
test: test-unit test-integration

test-unit:
	go test ./tests/unit/... -v

test-integration:
	go test ./tests/integration/... -v

test-with-db: docker-up-db
	sleep 10
	go test ./tests/unit/... -v
	go test ./tests/integration/... -v
	docker-compose down arangodb

# Docker
docker-build:
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

docker-up:
	docker-compose up -d

docker-up-db:
	docker-compose up -d arangodb

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Development
fmt:
	go fmt ./...

lint:
	golangci-lint run

clean:
	go clean
	rm -f $(BINARY_NAME)

# Dependencies
deps:
	go mod download
	go mod tidy

# Run locally
run:
	go run ./cmd/$(APP_NAME)

# Environment setup
env:
	@echo "Setting up environment variables..."
	@echo "ARANGO_URL=http://localhost:8529"
	@echo "ARANGO_DATABASE=stag"
	@echo "ARANGO_USERNAME=root"
	@echo "ARANGO_PASSWORD=stagpassword"
	@echo "STAG_PORT=8080"
	@echo "LOG_LEVEL=info"