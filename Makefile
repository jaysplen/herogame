# herogame dev tooling — see docs/architecture.md §12

.PHONY: dev down logs ps migrate migrate-status config

# Load .env when present (copy from .env.example)
ifneq (,$(wildcard .env))
include .env
export
endif

DATABASE_URL ?= postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0
GOOSE ?= goose
GOOSE_DIR ?= backend/migrations

## dev: start Postgres + Redis, wait for health, then run migrations if available
dev:
	docker compose up -d --wait
	@$(MAKE) migrate

## down: stop containers (volumes preserved)
down:
	docker compose down

## down-v: stop containers and remove named volumes (destructive)
down-v:
	docker compose down -v

logs:
	docker compose logs -f

ps:
	docker compose ps

config:
	docker compose config

## migrate: apply SQL migrations via goose (requires ALPHA-002 backend/migrations)
migrate:
	@set -e; \
	if [ ! -d "$(GOOSE_DIR)" ]; then \
		echo "Skipping migrations: $(GOOSE_DIR) not found (complete ALPHA-002 first)."; \
		exit 0; \
	fi; \
	command -v $(GOOSE) >/dev/null 2>&1 || { \
		echo "goose not found. Install: go install github.com/pressly/goose/v3/cmd/goose@latest"; \
		exit 1; \
	}; \
	$(GOOSE) -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" up

sqlc:
	cd backend && sqlc generate

migrate-status:
	@command -v $(GOOSE) >/dev/null 2>&1 || { echo "goose not found"; exit 1; }
	$(GOOSE) -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" status
