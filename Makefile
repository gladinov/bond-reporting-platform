-include Makefile.local
-include Makefile.test

# =========================
# Variables
# =========================

COMPOSE_FILE = deployments/docker-compose.yml
ENV_FILE = deployments/env/prod.env

# =========================
# Phony targets
# =========================

.PHONY: up build-up down up-services down-services build-up-services
        
# =========================
# Docker: production
# =========================

up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d

build-up:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down --remove-orphans
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build

down:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v

up-services:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d moex_app cbr_app tinkoffapi_app bond-report-service gateway notification_service

down-services:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v moex_app cbr_app tinkoffapi_app bond-report-service gateway notification_service

build-up-services:
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) down -v moex_app cbr_app tinkoffapi_app bond-report-service gateway notification_service
	docker compose -f $(COMPOSE_FILE) --env-file $(ENV_FILE) up -d --build moex_app cbr_app tinkoffapi_app bond-report-service gateway notification_service
