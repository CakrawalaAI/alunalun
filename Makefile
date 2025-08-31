# Makefile for Alunalun project
# Each feature can add their commands in their designated section

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make dev          - Start development servers"
	@echo "  make build        - Build all services"
	@echo "  make test         - Run all tests"
	@echo "  make proto        - Generate proto files"
	@echo "  make sqlc         - Generate SQLC code"
	@echo "  make migrate      - Run database migrations"
	@echo "  make seed         - Seed database with test data"

# === DEVELOPMENT COMMANDS START ===

.PHONY: dev
dev:
	# Start all development servers
	@echo "Starting development servers..."
	# Add concurrent server starts here

# [POSTS-DEV-START]
# Owner: feature/posts-core
# Add posts-specific dev commands
# [POSTS-DEV-END]

# [COMMENTS-DEV-START]
# Owner: feature/comments-system
# Add comments-specific dev commands
# [COMMENTS-DEV-END]

# [FEED-DEV-START]
# Owner: feature/location-feed
# Add feed-specific dev commands
# [FEED-DEV-END]

# === DEVELOPMENT COMMANDS END ===

# === BUILD COMMANDS START ===

.PHONY: build
build: build-api build-web

.PHONY: build-api
build-api:
	@echo "Building API server..."
	cd api && go build -o bin/server cmd/server/main.go

.PHONY: build-web
build-web:
	@echo "Building web client..."
	cd web && bun run build

# === BUILD COMMANDS END ===

# === CODE GENERATION START ===

.PHONY: proto
proto:
	@echo "Generating proto files..."
	cd api && buf generate

.PHONY: sqlc
sqlc:
	@echo "Generating SQLC code..."
	cd api && sqlc generate

# === CODE GENERATION END ===

# === DATABASE COMMANDS START ===

.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	# Add migration command here

.PHONY: migrate-down
migrate-down:
	@echo "Rolling back last migration..."
	# Add rollback command here

# [POSTS-SEED-START]
# Owner: feature/posts-core
# Add posts seed data
# [POSTS-SEED-END]

# [COMMENTS-SEED-START]
# Owner: feature/comments-system
# Add comments seed data
# [COMMENTS-SEED-END]

# === DATABASE COMMANDS END ===

# === TEST COMMANDS START ===

.PHONY: test
test: test-api test-web

.PHONY: test-api
test-api:
	@echo "Running API tests..."
	cd api && go test ./...

.PHONY: test-web
test-web:
	@echo "Running web tests..."
	cd web && bun test

# [POSTS-TEST-START]
# Owner: feature/posts-core
# Add posts-specific tests
# [POSTS-TEST-END]

# [COMMENTS-TEST-START]
# Owner: feature/comments-system
# Add comments-specific tests
# [COMMENTS-TEST-END]

# === TEST COMMANDS END ===