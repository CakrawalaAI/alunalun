# API Makefile for Alunalun project
# Database and API-specific commands

# Build database URL from environment variables
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/dev_alunalun?sslmode=$(DB_SSLMODE)

.PHONY: help
help:
	@echo "Available API commands:"
	@echo "  make setup-dev    - Setup development database (create, migrate, seed with auto-recovery)"
	@echo "  make reset-dev    - Reset development database to clean slate"
	@echo "  make create-db    - Create development database only"
	@echo "  make migrate      - Run database migrations with goose"
	@echo "  make migrate-down - Rollback last migration"
	@echo "  make seed         - Seed database with development data"
	@echo "  make build        - Build API server"
	@echo "  make run          - Run API server"
	@echo "  make test         - Run API tests"
	@echo "  make proto        - Generate proto files"
	@echo "  make sqlc         - Generate SQLC code"

# === DATABASE COMMANDS START ===

.PHONY: setup-dev
setup-dev:
	@echo "🚀 Setting up development database with auto-recovery..."
	@$(MAKE) _create-db || (echo "❌ Database creation failed - running reset..." && $(MAKE) reset-dev && exit 1)
	@$(MAKE) _migrate || (echo "❌ Migration failed - running reset..." && $(MAKE) reset-dev && exit 1)
	@$(MAKE) _seed || (echo "❌ Seeding failed - running reset..." && $(MAKE) reset-dev && exit 1)
	@echo "✅ Development database setup completed successfully!"

.PHONY: _create-db
_create-db:
	@echo "🏗️  Creating database..."
	@go run scripts/create-db.go --yes

.PHONY: _migrate
_migrate:
	@echo "📊 Running migrations..."
	@goose -dir sql/migrations postgres "$(DB_URL)" up
	@echo "🗃️  Regenerating SQLC code after migrations..."
	@sqlc generate

.PHONY: _seed
_seed:
	@echo "🌱 Seeding data..."
	@go run scripts/seed.go --yes

.PHONY: reset-dev
reset-dev:
	@echo "🧹 Resetting development database to clean slate..."
	@echo "⚠️  This will DROP and recreate the entire database!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	@go run scripts/create-db.go --yes || true
	@psql "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/postgres?sslmode=$(DB_SSLMODE)" -c 'DROP DATABASE IF EXISTS "dev_alunalun"' 2>/dev/null || true
	@$(MAKE) _create-db
	@echo "✅ Database reset completed"

# Public commands for manual use
.PHONY: create-db
create-db:
	@echo "🏗️  Creating development database..."
	@go run scripts/create-db.go


.PHONY: migrate
migrate:
	@echo "📊 Running database migrations..."
	@goose -dir sql/migrations postgres "$(DB_URL)" up
	@echo "🗃️  Regenerating SQLC code after migrations..."
	@sqlc generate

.PHONY: migrate-down
migrate-down:
	@echo "⬇️  Rolling back last migration..."
	@goose -dir sql/migrations postgres "$(DB_URL)" down

.PHONY: seed
seed:
	@echo "🌱 Seeding development database..."
	@go run scripts/seed.go

# === DATABASE COMMANDS END ===

# === BUILD COMMANDS START ===

.PHONY: build
build:
	@echo "🔨 Building API server..."
	@go build -o bin/server cmd/api/main.go

.PHONY: run
run:
	@echo "🏃 Starting API server..."
	@go run cmd/api/main.go

# === BUILD COMMANDS END ===

# === CODE GENERATION START ===

.PHONY: proto
proto:
	@echo "📝 Generating proto files..."
	@buf generate

.PHONY: sqlc
sqlc:
	@echo "🗃️  Generating SQLC code..."
	@sqlc generate

# === CODE GENERATION END ===

# === TEST COMMANDS START ===

.PHONY: test
test:
	@echo "🧪 Running API tests..."
	@go test ./...

.PHONY: test-verbose
test-verbose:
	@echo "🧪 Running API tests (verbose)..."
	@go test -v ./...

# === TEST COMMANDS END ===