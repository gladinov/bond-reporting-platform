# =========================
# Variables
# =========================

COMPOSE_FILE = docker-compose.yml
ENV_FILE = ../../deployments/env/prod.env
GOLANGCI_LINT = $(shell go env GOPATH)/bin/golangci-lint
LINT_PACKAGES = ./cmd ./cmd/migrator ./internal/adapters/... ./internal/app ./internal/closer ./internal/configs ./internal/domain/... ./internal/migrator ./internal/utils/...

# =========================
# Phony targets
# =========================

.PHONY: up build-up down migrate build-up-app test test-integration lint

# =========================
# Docker: production
# =========================

up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d

build-up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v

migrate:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up --build bond-report-migrator

build-up-app:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v bond-report-service
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build bond-report-service

test:
	go test ./...

test-integration:
	go test -tags=integration ./...

lint:
	XDG_CACHE_HOME=/tmp GOLANGCI_LINT_CACHE=/tmp/golangci-lint $(GOLANGCI_LINT) run $(LINT_PACKAGES)
