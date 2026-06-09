.PHONY: build run test lint audit clean migrate

# Build
build:
	go build -o bin/devbrain ./cmd/server

# Run
run:
	go run ./cmd/server

# Test
test:
	go test -v -race -cover ./...

# Lint
lint:
	go vet ./...
	golangci-lint run

# Security audit
audit:
	govulncheck ./...

# Clean
clean:
	rm -rf bin/
	go clean

# Database migrations
migrate:
	migrate -path ./migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

# Development
dev:
	air

# Production build
prod:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/devbrain ./cmd/server

# Docker
docker-build:
	docker build -t devbrain:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env devbrain:latest

# Pre-release checks
pre-release: lint audit test build
	@echo "✓ All checks passed!"
