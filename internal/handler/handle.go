package handler

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Handler gRPC处理器
type Handler struct {
	logger *zap.Logger
}

// NewHandler 创建新的处理器
func NewHandler(
	logger *zap.Logger,
) *Handler {
	return &Handler{
		logger: logger,
	}
}

// RegisterServices 注册gRPC服务
func (h *Handler) RegisterServices(server *grpc.Server) {
	// 注册用户服务
	//userHandler := NewUserHandler(h.userService, h.logger)

	h.logger.Info("gRPC services registered successfully")

	// 注意：这里注释掉了实际的注册代码，因为需要先生成protobuf代码
	// 在实际使用时，需要：
	// 1. 安装protoc和相关插件
	// 2. 生成Go代码：protoc --go_out=. --go-grpc_out=. proto/flashcard.proto
	// 3. 取消注释上面的注册代码
}
