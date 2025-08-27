# fuck-comment 构建配置

# 项目信息
BINARY_NAME=fuck-comment
VERSION=1.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go 构建参数
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
.PHONY: all
all: clean build

# 清理构建文件
.PHONY: clean
clean:
	rm -rf dist/
	go clean

# 安装依赖
.PHONY: deps
deps:
	go mod tidy
	go mod download

# 本地构建
.PHONY: build
build: deps
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# 跨平台构建
.PHONY: build-all
build-all: clean deps
	mkdir -p dist
	
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-386.exe .
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-arm64.exe .
	
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=386 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-386 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .
	GOOS=linux GOARCH=arm go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm .

# 运行测试
.PHONY: test
test:
	go test -v ./...

# 安装到本地
.PHONY: install
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# 卸载
.PHONY: uninstall
uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

# 显示帮助
.PHONY: help
help:
	@echo "fuck-comment 构建工具"
	@echo ""
	@echo "可用命令:"
	@echo "  make build      - 构建当前平台版本"
	@echo "  make build-all  - 构建所有平台版本"
	@echo "  make clean      - 清理构建文件"
	@echo "  make deps       - 安装依赖"
	@echo "  make test       - 运行测试"
	@echo "  make install    - 安装到系统"
	@echo "  make uninstall  - 从系统卸载"
	@echo "  make help       - 显示此帮助信息"
