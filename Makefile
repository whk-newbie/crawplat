.PHONY: test docker-binaries web-assets migrate up down dev-up dev-down

COMPOSE_FILE := deploy/docker-compose/docker-compose.mcp.yml
DEV_COMPOSE_FILE := deploy/docker-compose/docker-compose.dev.yml
DOCKER_BIN_DIR := .docker-bin

NESTED_GO_MODULES := $(shell find packages apps -name go.mod -exec dirname {} \; | sort)

test:
	go test ./...
	@set -e; \
	for mod in $(NESTED_GO_MODULES); do \
		( cd "$$mod" && go test ./... ); \
	done
	( cd apps/web && npm test )

docker-binaries:
	mkdir -p $(DOCKER_BIN_DIR)
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/gateway ./apps/gateway/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/iam-service ./apps/iam-service/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/project-service ./apps/project-service/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/spider-service ./apps/spider-service/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/execution-service ./apps/execution-service/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/node-service ./apps/node-service/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/datasource-service ./apps/datasource-service/cmd/server
	CGO_ENABLED=0 GOOS=linux go build -o $(DOCKER_BIN_DIR)/agent ./apps/agent/cmd/agent

web-assets:
	npm --prefix apps/web run build

migrate:
	./deploy/scripts/migrate-postgres.sh

up: docker-binaries web-assets migrate
	docker compose -f $(COMPOSE_FILE) up --build -d

down:
	docker compose -f $(COMPOSE_FILE) down -v --remove-orphans

dev-up:
	COMPOSE_FILE=$(PWD)/$(DEV_COMPOSE_FILE) ./deploy/scripts/migrate-postgres.sh
	docker compose -f $(DEV_COMPOSE_FILE) up --build -d

dev-down:
	docker compose -f $(DEV_COMPOSE_FILE) down -v --remove-orphans
