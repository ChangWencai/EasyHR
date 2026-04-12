# EasyHR Makefile
# 易人事 - 小微企业人事管理APP

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# =============================================================================
# Variables
# =============================================================================
GO_CMD      := github.com/wencai/easyhr/cmd/server
GO_BIN      := easyhr
GO_PORT     := 8080
FE_PORT     := 5173
DOCKER_COMPOSE := docker-compose

# =============================================================================
# Go Backend
# =============================================================================
.PHONY: go/build go/run go/test go/vet go/fmt go/clean
go/build: ## Build Go binary to ./server
	CGO_ENABLED=0 go build -o ./server $(GO_CMD)

go/run: ## Run Go server (requires dependencies up)
	./server

go/test: ## Run Go tests
	go test -race -cover ./...

go/vet: ## Run go vet
	go vet ./...

go/fmt: ## Format Go code
	go fmt ./...
	gofmt -w .

go/clean: ## Remove build artifacts
	rm -f ./server
	rm -f ./easyhr

# =============================================================================
# Frontend (Vue 3 H5)
# =============================================================================
.PHONY: fe/deps fe/dev fe/build fe/type-check fe/lint fe/test
FE_DIR := frontend

fe/deps: ## Install frontend npm dependencies
	cd $(FE_DIR) && npm install

fe/dev: ## Start Vite dev server on :5173
	cd $(FE_DIR) && npm run dev

fe/build: ## Build frontend to frontend/dist
	cd $(FE_DIR) && npm run build

fe/type-check: ## Run vue-tsc type check
	cd $(FE_DIR) && npm run type-check

fe/lint: ## Run ESLint with fix
	cd $(FE_DIR) && npm run lint

fe/test: ## Run vitest unit tests
	cd $(FE_DIR) && npm run test:unit

# =============================================================================
# Docker
# =============================================================================
.PHONY: docker/build docker/up docker/down docker/ps
docker/build: ## Build production Docker image
	docker build -t easyhr:latest .

docker/up: ## Start postgres + redis via docker-compose
	$(DOCKER_COMPOSE) up -d

docker/down: ## Stop docker-compose services
	$(DOCKER_COMPOSE) down

docker/ps: ## Show docker-compose status
	$(DOCKER_COMPOSE) ps

# =============================================================================
# Development
# =============================================================================
.PHONY: dev/up dev/down dev/restart dev/status
dev/up: docker/up ## Start all dev services (postgres, redis)
	@echo "Waiting for postgres..."
	@until docker exec $$(docker-compose ps -q postgres) pg_isready -U easyhr -d easyhr > /dev/null 2>&1; do sleep 1; done
	@echo "Waiting for redis..."
	@until docker exec $$(docker-compose ps -q redis) redis-cli ping > /dev/null 2>&1; do sleep 1; done
	@echo "Ready. Postgres :5433, Redis :6380, Server :$(GO_PORT), Frontend :$(FE_PORT)"

dev/down: docker/down ## Stop all dev services

dev/restart: docker/down docker/up ## Restart all dev services

dev/status: docker/ps ## Show dev service status

# =============================================================================
# Full Stack
# =============================================================================
.PHONY: up down logs
up: ## Start all services in background (docker infra + backend + frontend)
	@echo "Starting docker services..."
	$(MAKE) docker/up
	@echo "Waiting for postgres..."
	@until docker exec $$(docker-compose ps -q postgres) pg_isready -U easyhr -d easyhr > /dev/null 2>&1; do sleep 1; done
	@echo "Waiting for redis..."
	@until docker exec $$(docker-compose ps -q redis) redis-cli ping > /dev/null 2>&1; do sleep 1; done
	@echo "Building backend..."
	$(MAKE) go/build
	@echo "Starting backend on :$(GO_PORT)..."
	./server > .server.log 2>&1 &
	@echo "Starting frontend on :$(FE_PORT)..."
	cd $(FE_DIR) && npm run dev > .frontend.log 2>&1 &
	@echo ""
	@echo "All services started:"
	@echo "  Backend  http://localhost:$(GO_PORT)/api/v1"
	@echo "  Frontend http://localhost:$(FE_PORT)"
	@echo "  Postgres localhost:5433"
	@echo "  Redis    localhost:6380"
	@echo ""
	@echo "Logs:"
	@echo "  make logs          # tail all logs"
	@echo "  make logs-backend  # backend log only"
	@echo "  make logs-frontend # frontend log only"

restart: ## Restart backend + frontend dev servers
	@echo "Stopping servers..."
	@kill $$(pgrep -f "./server" 2>/dev/null) 2>/dev/null || true
	@kill $$(pgrep -f "vite" 2>/dev/null) 2>/dev/null || true
	@echo "Building backend..."
	$(MAKE) go/build
	@echo "Starting backend..."
	./server > .server.log 2>&1 &
	@echo "Starting frontend..."
	cd $(FE_DIR) && npm run dev > .frontend.log 2>&1 &
	@echo "Done. Backend :$(GO_PORT), Frontend :$(FE_PORT)"

down: ## Stop all services (docker + backend + frontend)
	@echo "Stopping all services..."
	@kill $$(pgrep -f "./server" 2>/dev/null) 2>/dev/null || true
	@kill $$(pgrep -f "vite" 2>/dev/null) 2>/dev/null || true
	$(MAKE) docker/down
	@echo "Done."

logs: ## Tail all service logs
	@echo "--- backend ---" && tail -f .server.log
logs-backend: ## Tail backend log
	@tail -f .server.log
logs-frontend: ## Tail frontend log
	@tail -f .frontend.log

# =============================================================================
# Production
# =============================================================================
.PHONY: prod/build prod/start
prod/build: go/build ## Build Go binary for production

prod/start: ## Run production binary (requires config/config.yaml)
	./server

# =============================================================================
# Utilities
# =============================================================================
.PHONY: clean lint test
clean: go/clean ## Clean all build artifacts (backend binary + frontend dist)
	rm -rf $(FE_DIR)/dist
	rm -rf $(FE_DIR)/node_modules

lint: go/fmt fe/lint ## Format and lint all code

test: go/test fe/test ## Run all tests (Go + frontend)
