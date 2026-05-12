SHELL := /bin/bash

DC := docker compose
DC_FILE := docker-compose.yaml

# =========================
# HELP
# =========================
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make up                - start full stack"
	@echo "  make down              - stop full stack"
	@echo "  make restart           - restart full stack"
	@echo "  make logs              - tail logs"
	@echo "  make ps                - show containers"
	@echo ""
	@echo "  make build             - build all services"
	@echo "  make rebuild           - rebuild without cache"
	@echo ""
	@echo "  make up-infra          - start only infra (db, kafka, redis, etc)"
	@echo "  make up-app            - start only app services"
	@echo ""
	@echo "  make gen-proto         - generate protobuf (Go)"
	@echo "  make gen-swagger 		- generate Swagger for api-gateway
	@echo ""
	@echo "  make restart-service S=service_name - restart single service"
	@echo "  make logs-service S=service_name    - logs for service"
	@echo ""
	@echo "  make clean             - remove volumes + containers"
	@echo "  make clean-restart		- remove volumes + containers & start up"

# =========================
# CORE LIFECYCLE
# =========================
.PHONY: up down restart ps logs

up:
	$(DC) -f $(DC_FILE) up -d

down:
	$(DC) -f $(DC_FILE) down

restart:
	$(DC) -f $(DC_FILE) down && $(DC) -f $(DC_FILE) up -d

ps:
	$(DC) -f $(DC_FILE) ps

logs:
	$(DC) -f $(DC_FILE) logs -f --tail=200

# =========================
# BUILD
# =========================
.PHONY: build rebuild

build:
	$(DC) -f $(DC_FILE) build

rebuild:
	$(DC) -f $(DC_FILE) build --no-cache

# =========================
# INFRA ONLY
# =========================
.PHONY: up-infra up-app

up-infra:
	$(DC) -f $(DC_FILE) up -d postgres redis kafka zookeeper kafka-init ray-head

up-app:
	$(DC) -f $(DC_FILE) up -d user_service experiment_service model_registry_service model_serving_service traffic_splitter api-gateway

# =========================
# SINGLE SERVICE CONTROL
# =========================
.PHONY: restart-service logs-service build-service

restart-service:
	$(DC) -f $(DC_FILE) restart $(S)

logs-service:
	$(DC) -f $(DC_FILE) logs -f --tail=200 $(S)

build-service:
	$(DC) -f $(DC_FILE) build $(S)

# =========================
# PROTO (Go)
# =========================
PROTO_DIR := protos
GEN_DIR := gen/go

.PHONY: gen-proto
gen-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GEN_DIR) \
		--go-grpc_out=$(GEN_DIR) \
		$(shell find $(PROTO_DIR) -name "*.proto")

# =========================
# SWAGGER
# =========================
.PHONY: gen-swagger
gen-swagger:
	swag init -g services/api_gateway/cmd/server/main.go -o services/api_gateway/docs

# =========================
# CLEAN
# =========================
.PHONY: clean clean-volumes

clean:
	$(DC) -f $(DC_FILE) down -v --remove-orphans

clean-restart:
	make clean && make up

clean-volumes:
	docker volume prune -f
