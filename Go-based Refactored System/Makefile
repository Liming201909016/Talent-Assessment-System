# Talent Assessment - Go Backend Makefile
# 用法: make <target>

.PHONY: help build test test-pkg test-handler lint coverage clean run

help:
	@echo "可用目标:"
	@echo "  make build        - 编译 server (输出到 bin/)"
	@echo "  make test         - 运行所有测试"
	@echo "  make test-pkg     - 仅测试 pkg/ 下的工具包"
	@echo "  make test-handler - 仅测试 handler 层"
	@echo "  make coverage     - 生成覆盖率报告 (HTML)"
	@echo "  make lint         - 运行 golangci-lint"
	@echo "  make clean        - 清理 bin/ 和 coverage 文件"
	@echo "  make run          - 启动本地开发服务器"

build:
	@echo "==> Building server..."
	@go build -o bin/server.exe ./cmd/server
	@echo "==> Built: bin/server.exe"

test:
	@echo "==> Running all tests..."
	@go test ./... -v -count=1

test-pkg:
	@go test ./pkg/... -v -count=1

test-handler:
	@go test ./internal/handler/... -v -count=1

coverage:
	@echo "==> Generating coverage report..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "==> HTML report: coverage.html"

lint:
	@echo "==> Running golangci-lint..."
	@golangci-lint run ./...

clean:
	@if exist bin rmdir /s /q bin
	@if exist coverage.out del coverage.out
	@if exist coverage.html del coverage.html

run:
	@set APP_ENV=local&& go run ./cmd/server
