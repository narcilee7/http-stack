# Go HTTP协议栈 Makefile

# 变量定义
GO_VERSION := 1.21
BINARY_NAME_SERVER := http-server
BINARY_NAME_CLIENT := http-client
BUILD_DIR := ./bin
CMD_SERVER_DIR := ./cmd/server
CMD_CLIENT_DIR := ./cmd/client
COVERAGE_FILE := coverage.txt
COVERAGE_HTML := coverage.html

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "HTTP协议栈开发工具"
	@echo ""
	@echo "可用命令："
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# 开发工具
.PHONY: setup
setup: ## 安装开发依赖
	@echo "安装开发工具..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "开发环境设置完成"

.PHONY: deps
deps: ## 下载Go模块依赖
	@echo "下载依赖..."
	go mod download
	go mod verify

.PHONY: tidy
tidy: ## 整理Go模块依赖
	@echo "整理依赖..."
	go mod tidy

# 代码质量
.PHONY: fmt
fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	goimports -w .

.PHONY: lint
lint: ## 运行代码检查
	@echo "运行代码检查..."
	golangci-lint run

.PHONY: vet
vet: ## 运行go vet
	@echo "运行vet检查..."
	go vet ./...

# 测试
.PHONY: test
test: ## 运行所有测试
	@echo "运行测试..."
	go test -race -v ./...

.PHONY: test-short
test-short: ## 运行短测试
	@echo "运行短测试..."
	go test -short -v ./...

.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "生成覆盖率报告..."
	go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "覆盖率报告生成: $(COVERAGE_HTML)"

.PHONY: test-coverage-func
test-coverage-func: ## 显示函数级覆盖率
	@echo "显示函数覆盖率..."
	go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -func=$(COVERAGE_FILE)

.PHONY: bench
bench: ## 运行基准测试
	@echo "运行基准测试..."
	go test -bench=. -benchmem ./...

.PHONY: bench-cpu
bench-cpu: ## 运行基准测试并生成CPU性能分析
	@echo "生成CPU性能分析..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof ./...

.PHONY: bench-mem
bench-mem: ## 运行基准测试并生成内存性能分析
	@echo "生成内存性能分析..."
	go test -bench=. -benchmem -memprofile=mem.prof ./...

# 构建
.PHONY: build
build: build-server build-client ## 构建所有二进制文件

.PHONY: build-server
build-server: ## 构建HTTP服务器
	@echo "构建服务器..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME_SERVER) $(CMD_SERVER_DIR)

.PHONY: build-client
build-client: ## 构建HTTP客户端
	@echo "构建客户端..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT) $(CMD_CLIENT_DIR)

.PHONY: build-race
build-race: ## 构建带竞态检测的二进制文件
	@echo "构建带竞态检测的二进制文件..."
	@mkdir -p $(BUILD_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME_SERVER)-race $(CMD_SERVER_DIR)
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME_CLIENT)-race $(CMD_CLIENT_DIR)

.PHONY: install
install: ## 安装二进制文件到GOPATH/bin
	@echo "安装二进制文件..."
	go install $(CMD_SERVER_DIR)
	go install $(CMD_CLIENT_DIR)

# 运行
.PHONY: run-server
run-server: ## 运行HTTP服务器
	@echo "启动HTTP服务器..."
	go run $(CMD_SERVER_DIR)

.PHONY: run-client
run-client: ## 运行HTTP客户端
	@echo "启动HTTP客户端..."
	go run $(CMD_CLIENT_DIR)

# 清理
.PHONY: clean
clean: ## 清理构建文件和临时文件
	@echo "清理文件..."
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	rm -f cpu.prof mem.prof
	rm -f *.test
	go clean -testcache

.PHONY: clean-cache
clean-cache: ## 清理Go缓存
	@echo "清理Go缓存..."
	go clean -cache
	go clean -modcache

# CI/CD相关
.PHONY: ci
ci: fmt lint vet test-coverage ## 运行CI检查流程

.PHONY: pre-commit
pre-commit: fmt lint vet test ## 提交前检查

.PHONY: check
check: ## 完整检查（格式化、lint、测试、构建）
	@echo "运行完整检查..."
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) vet
	$(MAKE) test
	$(MAKE) build

# 文档
.PHONY: doc
doc: ## 生成并打开Go文档
	@echo "生成文档..."
	godoc -http=:6060 &
	@echo "文档服务器启动在 http://localhost:6060"

# 开发辅助
.PHONY: todo
todo: ## 显示代码中的TODO项
	@echo "查找TODO项..."
	@grep -r "TODO\|FIXME\|XXX" --include="*.go" . || echo "没有找到TODO项"

.PHONY: stats
stats: ## 显示代码统计
	@echo "代码统计:"
	@find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | tail -1
	@echo "文件数量:"
	@find . -name "*.go" -not -path "./vendor/*" | wc -l

# 性能分析
.PHONY: pprof-cpu
pprof-cpu: ## 查看CPU性能分析
	go tool pprof cpu.prof

.PHONY: pprof-mem
pprof-mem: ## 查看内存性能分析
	go tool pprof mem.prof

# 安全检查
.PHONY: security
security: ## 运行安全扫描
	@echo "运行安全扫描..."
	gosec ./...

# 依赖检查
.PHONY: deps-check
deps-check: ## 检查依赖更新
	@echo "检查依赖更新..."
	go list -u -m all

.PHONY: deps-upgrade
deps-upgrade: ## 升级依赖
	@echo "升级依赖..."
	go get -u ./...
	go mod tidy

# Git hooks
.PHONY: install-hooks
install-hooks: ## 安装Git hooks
	@echo "安装Git hooks..."
	@echo '#!/bin/sh\nmake pre-commit' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Git hooks安装完成"

# 快速目标
.PHONY: quick
quick: fmt test ## 快速检查（格式化+测试）

.PHONY: all
all: check ## 运行所有检查和构建
