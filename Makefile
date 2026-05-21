# herogame dev tooling — see docs/architecture.md §12

.PHONY: dev down logs ps migrate migrate-status config server frontend

# Load .env when present (copy from .env.example)
ifneq (,$(wildcard .env))
include .env
export
endif

DATABASE_URL ?= postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable
REDIS_URL ?= redis://localhost:6379/0
GOOSE_DIR ?= backend/migrations
# Real user home when invoked via sudo (goose is installed in ~user/go/bin, not /root)
REAL_HOME := $(if $(SUDO_USER),$(shell getent passwd $(SUDO_USER) 2>/dev/null | cut -d: -f6),$(HOME))
GOOSE_PATHS := $(REAL_HOME)/go/bin/goose $(shell go env GOPATH 2>/dev/null)/bin/goose
GOOSE_BIN := $(firstword $(shell command -v goose 2>/dev/null) $(foreach p,$(GOOSE_PATHS),$(wildcard $(p))))
export PATH := $(REAL_HOME)/go/bin:$(shell go env GOPATH 2>/dev/null)/bin:$(PATH)

## server: run game server on :8080 (requires make dev for Postgres/Redis)
server:
	bash scripts/dev-backend.sh

## frontend: Vite dev server on :5173
frontend:
	cd frontend && npm run dev -- --host 0.0.0.0

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
	run_goose() { \
		if [ -n "$(GOOSE_BIN)" ] && [ -x "$(GOOSE_BIN)" ]; then \
			"$(GOOSE_BIN)" "$$@"; \
		elif command -v goose >/dev/null 2>&1; then \
			goose "$$@"; \
		elif command -v go >/dev/null 2>&1; then \
			go run github.com/pressly/goose/v3/cmd/goose@latest "$$@"; \
		else \
			echo "goose not found and go not on PATH."; \
			echo "Install: go install github.com/pressly/goose/v3/cmd/goose@latest"; \
			echo "Add \$$(go env GOPATH)/bin to PATH, or run make dev without sudo."; \
			exit 1; \
		fi; \
	}; \
	run_goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" up

sqlc:
	cd backend && sqlc generate

migrate-status:
	@set -e; \
	run_goose() { \
		if [ -n "$(GOOSE_BIN)" ] && [ -x "$(GOOSE_BIN)" ]; then "$(GOOSE_BIN)" "$$@"; \
		elif command -v goose >/dev/null 2>&1; then goose "$$@"; \
		elif command -v go >/dev/null 2>&1; then go run github.com/pressly/goose/v3/cmd/goose@latest "$$@"; \
		else echo "goose/go not found"; exit 1; fi; \
	}; \
	run_goose -dir $(GOOSE_DIR) postgres "$(DATABASE_URL)" status
