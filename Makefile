# EasyHR Makefile
# 用法: make <target>
#
# 前置要求:
#   - Go 1.23+
#   - Node.js 18+ & npm
#   - Docker & Docker Compose (用于启动数据库/Redis)
#
# 基础设施（PostgreSQL / Redis）:
#   make infra-up     启动数据库和 Redis（docker-compose）
#   make infra-down   停止基础设施
#   make infra-logs   查看基础设施日志
#
# 开发:
#   make init         安装前端依赖（首次克隆后运行一次）
#   make dev          同时启动前端 + 后端（热重载）
#   make dev-backend  仅启动后端（热重载）
#   make dev-frontend 仅启动前端开发服务器
#
# 构建:
#   make build              构建前端（输出到 frontend/dist）
#   make build-backend      构建后端二进制
#
# 数据库:
#   make db-migrate    运行数据库迁移（AutoMigrate）
#   make db-reset      重置数据库（删除数据后重建）
#
# 清理:
#   make clean         删除前端构建产物
#   make clean-all     删除构建产物 + node_modules
#
# 测试:
#   make test          运行所有测试
#   make test-wxmp     仅运行小程序模块测试
#
# ============================================================

.DEFAULT_GOAL := help

# 项目路径
ROOT          := $(shell pwd)
FRONTEND_DIR  := $(ROOT)/frontend
BACKEND_DIR   := $(ROOT)
CMD_DIR       := $(BACKEND_DIR)/cmd/server
BACKEND_BIN   := $(ROOT)/bin/server

# Go / npm
GO      := go
GOBUILD := $(GO) build -v
GOTEST  := $(GO) test -v
NPM     := npm

# Docker Compose
DC      := docker compose
DC_FILE := $(ROOT)/docker-compose.yml

# ============================================================
# 颜色（直接赋值，终端支持时生效）
# ============================================================

ESC := \033
BOLD    := $(shell printf '$(ESC)[1m' 2>/dev/null || printf '')
GREEN_  := $(shell printf '$(ESC)[32m' 2>/dev/null || printf '')
YELLOW_ := $(shell printf '$(ESC)[33m' 2>/dev/null || printf '')
CYAN_   := $(shell printf '$(ESC)[36m' 2>/dev/null || printf '')
RESET_  := $(shell printf '$(ESC)[0m' 2>/dev/null || printf '')

# ============================================================
# 基础设施
# ============================================================

.PHONY: infra-up infra-down infra-restart infra-logs

infra-up: ## 启动 PostgreSQL 和 Redis
	@printf "%s%s==> 启动基础设施 (PostgreSQL :5432, Redis :6379)...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	$(DC) -f $(DC_FILE) up -d
	@printf "%s%s==> 基础设施已启动%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"
	@printf "  PostgreSQL: localhost:5432 (user: easyhr, pass: easyhr_dev, db: easyhr)\n"
	@printf "  Redis:      localhost:6380\n"

infra-down: ## 停止 PostgreSQL 和 Redis
	@printf "%s%s==> 停止基础设施...%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	$(DC) -f $(DC_FILE) down

infra-restart: infra-down infra-up ## 重启基础设施

infra-logs: ## 查看基础设施日志
	$(DC) -f $(DC_FILE) logs -f

# ============================================================
# 初始化
# ============================================================

.PHONY: init

init: $(FRONTEND_DIR)/node_modules ## 安装前端依赖（首次运行一次）

$(FRONTEND_DIR)/node_modules: $(FRONTEND_DIR)/package.json
	@printf "%s%s==> 安装前端依赖 (npm install)...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(FRONTEND_DIR) && $(NPM) install
	@printf "%s%s==> 前端依赖安装完成%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

# ============================================================
# 开发
# ============================================================

.PHONY: dev dev-backend dev-frontend

dev: init ## 启动前端 + 后端开发服务器（并行）
	@printf "%s%s==> 启动前后端开发服务器...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	@printf "%s%s    后端: http://localhost:8089%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	@printf "%s%s    前端: http://localhost:5173%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	@printf "%s%s    小程序: 微信开发者工具打开 miniprogram/%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	trap 'kill 0' INT; \
		$(MAKE) dev-backend & \
		$(MAKE) dev-frontend & \
		wait

dev-backend: ## 启动后端热重载开发服务器
	@printf "%s%s==> 启动后端 (localhost:8089)...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(BACKEND_DIR) && $(GO) run $(CMD_DIR)/main.go

dev-frontend: init ## 启动前端热重载开发服务器
	@printf "%s%s==> 启动前端开发服务器 (localhost:5173)...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(FRONTEND_DIR) && $(NPM) run dev

# ============================================================
# 构建
# ============================================================

.PHONY: build build-backend build-frontend clean clean-all

build: build-frontend build-backend ## 构建前端 + 后端

run: build-frontend ## 构建前端并运行后端
	@printf "%s%s==> 启动后端...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	@printf "%s%s    后端: http://localhost:8089%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	@printf "%s%s    前端: http://localhost:5173 (已构建至 frontend/dist)%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	cd $(BACKEND_DIR) && $(GO) run $(CMD_DIR)/main.go

build-frontend: ## 构建前端生产版本 (输出: frontend/dist)
	@printf "%s%s==> 构建前端...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(FRONTEND_DIR) && $(NPM) run build
	@printf "%s%s==> 前端构建完成: frontend/dist/%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

build-backend: ## 构建后端二进制 (输出: bin/server)
	@printf "%s%s==> 构建后端...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	mkdir -p $(ROOT)/bin
	cd $(BACKEND_DIR) && $(GOBUILD) -o $(BACKEND_BIN) $(CMD_DIR)/main.go
	@printf "%s%s==> 后端构建完成: %s%s\n" "$(BOLD)" "$(GREEN_)" "$(BACKEND_BIN)" "$(RESET_)"

# ============================================================
# 数据库
# ============================================================

.PHONY: db-migrate db-reset

db-migrate: ## 运行数据库迁移 (AutoMigrate)
	@printf "%s%s==> 运行数据库迁移...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(BACKEND_DIR) && $(GO) run $(CMD_DIR)/main.go
	@printf "%s%s==> 迁移完成%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

db-reset: infra-down ## 重置数据库（删除所有数据！）
	@printf "%s%s==> 重置数据库...%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	$(DC) -f $(DC_FILE) down -v 2>/dev/null || true
	$(MAKE) infra-up
	@printf "%s%s==> 数据库已重置%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

# ============================================================
# 测试
# ============================================================

.PHONY: test test-wxmp test-coverage

test: ## 运行所有测试
	@printf "%s%s==> 运行测试...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(BACKEND_DIR) && $(GOTEST) ./...

test-wxmp: ## 仅运行小程序模块测试
	@printf "%s%s==> 运行 wxmp 模块测试...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(BACKEND_DIR) && $(GOTEST) ./internal/wxmp/...

test-coverage: ## 运行测试并输出覆盖率
	@printf "%s%s==> 运行测试（含覆盖率）...%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	cd $(BACKEND_DIR) && $(GO) test -cover -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@printf "%s%s==> 覆盖率报告: coverage.html%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

# ============================================================
# 清理
# ============================================================

clean: ## 删除前端构建产物
	@printf "%s%s==> 清理构建产物...%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(BACKEND_BIN)
	rm -rf $(BACKEND_DIR)/coverage.out $(BACKEND_DIR)/coverage.html
	@printf "%s%s==> 清理完成%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

clean-all: clean ## 删除构建产物 + node_modules
	@printf "%s%s==> 深度清理（包含 node_modules）...%s\n" "$(BOLD)" "$(YELLOW_)" "$(RESET_)"
	rm -rf $(FRONTEND_DIR)/node_modules
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(BACKEND_BIN)
	rm -rf $(BACKEND_DIR)/coverage.out $(BACKEND_DIR)/coverage.html
	@printf "%s%s==> 深度清理完成%s\n" "$(BOLD)" "$(GREEN_)" "$(RESET_)"

# ============================================================
# 帮助
# ============================================================

.PHONY: help
help:
	@printf "\n"
	@printf "%s%s  EasyHR Makefile%s\n" "$(BOLD)" "$(CYAN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s基础设施:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %smake infra-up%s       启动 PostgreSQL (:5432) 和 Redis (:6379)\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake infra-down%s     停止基础设施\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake infra-logs%s     查看基础设施日志\n" "$(GREEN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s开发:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %smake init%s           安装前端依赖（首次运行一次）\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake dev%s            启动前端 + 后端开发服务器\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake dev-backend%s    仅启动后端\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake dev-frontend%s   仅启动前端\n" "$(GREEN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s构建:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %smake build%s          构建前端 + 后端\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake build-frontend%s 构建前端\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake build-backend%s  构建后端\n" "$(GREEN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s数据库:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %smake db-migrate%s     运行数据库迁移\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake db-reset%s       重置数据库（删除所有数据）\n" "$(GREEN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s测试:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %smake test%s           运行所有测试\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake test-wxmp%s      仅运行小程序模块测试\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake test-coverage%s  运行测试并生成覆盖率报告\n" "$(GREEN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s清理:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %smake clean%s          删除构建产物\n" "$(GREEN_)" "$(RESET_)"
	@printf "  %smake clean-all%s     删除构建产物 + node_modules\n" "$(GREEN_)" "$(RESET_)"
	@printf "\n"
	@printf "%s常用流程:%s\n" "$(BOLD)" "$(RESET_)"
	@printf "  %s# 首次克隆后:%s\n" "$(YELLOW_)" "$(RESET_)"
	@printf "    make init\n"
	@printf "\n"
	@printf "  %s# 日常开发:%s\n" "$(YELLOW_)" "$(RESET_)"
	@printf "    make infra-up   # 启动数据库\n"
	@printf "    make dev        # 启动前后端\n"
	@printf "\n"
	@printf "  %s# 微信小程序调试:%s\n" "$(YELLOW_)" "$(RESET_)"
	@printf "    # 1. 打开微信开发者工具，导入 miniprogram/\n"
	@printf "    # 2. 后端: make dev-backend\n"
	@printf "    # 3. 开发者工具【详情 → 本地设置】勾选【不校验合法域名】\n"
	@printf "\n"
