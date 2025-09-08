# Makefile for Flashcard Backend

.PHONY: help build run test clean proto deps

# 默认目标
help:
	@echo "Available commands:"
	@echo "  build    - Build the application"
	@echo "  run      - Run the application"
	@echo "  test     - Run tests"
	@echo "  clean    - Clean build artifacts"
	@echo "  proto    - Generate protobuf code"
	@echo "  deps     - Download dependencies"
	@echo "  dev      - Run in development mode"

# 构建应用
build:
	@echo "Building flashcard-backend..."
	go build -o bin/server cmd/server/main.go

# 运行应用
run: build
	@echo "Starting flashcard-backend server..."
	./bin/server

# 开发模式运行
dev:
	@echo "Running in development mode..."
	go run cmd/server/main.go

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 清理构建产物
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf proto/flashcard/*.pb.go

# 生成protobuf代码
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go-grpc_out=. proto/flashcard.proto

# 下载依赖
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 安装开发工具
tools:
	@echo "Installing development tools..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 创建数据库
db-create:
	@echo "Creating database..."
	createdb flashcard_db

# 删除数据库
db-drop:
	@echo "Dropping database..."
	dropdb flashcard_db