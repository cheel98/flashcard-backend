package app

import (
	"fmt"
	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/cheel98/flashcard-backend/internal/handler"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server 服务器结构体
type Server struct {
	config     *config.Config
	logger     *zap.Logger
	grpcServer *grpc.Server
	handler    *handler.Handler
}

// NewServer 创建新的服务器实例
func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	handler *handler.Handler,
) *Server {
	grpcServer := grpc.NewServer()

	return &Server{
		config:     cfg,
		logger:     logger,
		grpcServer: grpcServer,
		handler:    handler,
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 注册gRPC服务
	s.handler.RegisterServices(s.grpcServer)

	// 监听端口
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		s.logger.Error("Failed to listen", zap.Error(err))
		return err
	}

	s.logger.Info("Server starting", zap.Int("port", s.config.Server.Port))

	// 启动gRPC服务器
	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			s.logger.Error("Failed to serve gRPC", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.logger.Info("Stopping server...")
	s.grpcServer.GracefulStop()
	return nil
}
