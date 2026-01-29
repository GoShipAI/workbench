# 工作台 Makefile
# Wails v2 项目常用命令

.PHONY: dev build build-all build-mac build-universal build-windows build-linux build-all-platforms clean run install frontend backend lint test help

# 默认目标
.DEFAULT_GOAL := help

# 变量
APP_NAME := Workbench
BUILD_DIR := build/bin
FRONTEND_DIR := frontend

# 开发模式（热重载）
dev:
	wails dev

# 构建当前平台版本
build:
	wails build

# 构建所有平台版本 (macOS + Windows)
build-all: build-mac build-windows
	@echo "构建完成！"
	@echo "macOS: $(BUILD_DIR)/$(APP_NAME).app"
	@echo "Windows: $(BUILD_DIR)/$(APP_NAME).exe"

# 构建 macOS 通用版本 (Intel + Apple Silicon)
build-mac:
	wails build -platform darwin/universal
	@echo "macOS 版本构建完成"

# 构建 macOS 通用版本 (别名)
build-universal: build-mac

# 构建 Windows 版本
build-windows:
	wails build -platform windows/amd64
	@echo "Windows 版本构建完成"

# 构建 Linux 版本
build-linux:
	wails build -platform linux/amd64
	@echo "Linux 版本构建完成"

# 构建所有平台版本 (macOS + Windows + Linux)
build-all-platforms: build-mac build-windows build-linux
	@echo "所有平台构建完成！"

# 清理构建产物
clean:
	rm -rf $(BUILD_DIR)/*
	rm -rf $(FRONTEND_DIR)/dist
	rm -rf $(FRONTEND_DIR)/node_modules/.vite

# 运行已构建的应用
run:
	open "$(BUILD_DIR)/$(APP_NAME).app"

# 安装前端依赖
frontend-install:
	cd $(FRONTEND_DIR) && npm install

# 更新前端依赖
frontend-update:
	cd $(FRONTEND_DIR) && npm update

# 前端类型检查
frontend-typecheck:
	cd $(FRONTEND_DIR) && npm run type-check 2>/dev/null || npx vue-tsc --noEmit

# 后端依赖更新
backend-update:
	go get -u ./...
	go mod tidy

# Go 代码格式化
fmt:
	go fmt ./...

# Go 代码检查
lint:
	go vet ./...

# 生成 Wails 绑定
bindings:
	wails generate module

# 查看应用日志 (macOS)
logs:
	log stream --predicate 'processImagePath contains "$(APP_NAME)"' --level debug

# 打开数据目录 (macOS)
open-data:
	open ~/Library/Application\ Support/$(APP_NAME)

# 查看数据库内容
db-info:
	@echo "Database location: ~/Library/Application Support/$(APP_NAME)/workbench.db"
	@sqlite3 ~/Library/Application\ Support/$(APP_NAME)/workbench.db ".tables" 2>/dev/null || echo "Database not found"

# 统计代码行数
loc:
	@echo "=== Go 代码 ==="
	@find . -name "*.go" -not -path "./frontend/*" | xargs wc -l | tail -1
	@echo "=== Vue 代码 ==="
	@find ./frontend/src -name "*.vue" -o -name "*.ts" | xargs wc -l | tail -1

# 帮助信息
help:
	@echo "工作台 开发命令"
	@echo ""
	@echo "开发:"
	@echo "  make dev              - 启动开发模式（热重载）"
	@echo "  make build            - 构建当前平台版本"
	@echo "  make build-all        - 构建 macOS + Windows 版本"
	@echo "  make build-mac        - 构建 macOS 通用版本"
	@echo "  make build-windows    - 构建 Windows 版本"
	@echo "  make build-linux      - 构建 Linux 版本"
	@echo "  make build-all-platforms - 构建所有平台版本"
	@echo "  make run              - 运行已构建的应用"
	@echo ""
	@echo "依赖:"
	@echo "  make frontend-install - 安装前端依赖"
	@echo "  make frontend-update  - 更新前端依赖"
	@echo "  make backend-update   - 更新后端依赖"
	@echo ""
	@echo "代码质量:"
	@echo "  make fmt              - 格式化 Go 代码"
	@echo "  make lint             - 检查 Go 代码"
	@echo "  make frontend-typecheck - 前端类型检查"
	@echo ""
	@echo "工具:"
	@echo "  make clean            - 清理构建产物"
	@echo "  make bindings         - 重新生成 Wails 绑定"
	@echo "  make logs             - 查看应用日志"
	@echo "  make open-data        - 打开数据目录"
	@echo "  make db-info          - 查看数据库信息"
	@echo "  make loc              - 统计代码行数"
