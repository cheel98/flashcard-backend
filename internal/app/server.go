package app

import (
	"fmt"
	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/cheel98/flashcard-backend/internal/handler"
	"github.com/cheel98/flashcard-backend/internal/middleware"
	grpcOptimizer "github.com/cheel98/flashcard-backend/internal/grpc"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server 服务器结构体
type Server struct {
	config          *config.Config
	logger          *zap.Logger
	grpcServer      *grpc.Server
	handler         *handler.Handler
	connectionPool  *grpcOptimizer.ConnectionPool
	workerPool      *grpcOptimizer.WorkerPool
	protobufOptimizer *grpcOptimizer.ProtobufOptimizer
	authMiddleware  *middleware.AuthMiddleware
}

// NewServer 创建新的服务器实例
func NewServer(
	cfg *config.Config,
	logger *zap.Logger,
	handler *handler.Handler,
	authMiddleware *middleware.AuthMiddleware,
) *Server {
	// 创建性能优化配置
	perfConfig := grpcOptimizer.DefaultPerformanceConfig()
	
	// 创建优化的gRPC服务器，集成JWT中间件
	grpcServer := grpcOptimizer.CreateOptimizedServerWithInterceptors(
		perfConfig,
		logger,
		authMiddleware.UnaryInterceptor(),
		authMiddleware.StreamInterceptor(),
	)
	
	// 创建连接池
	connPoolConfig := grpcOptimizer.DefaultConnectionPoolConfig()
	connectionPool := grpcOptimizer.NewConnectionPool(connPoolConfig, logger)
	
	// 创建工作池
	workerPool := grpcOptimizer.NewWorkerPool(perfConfig.WorkerPoolSize, perfConfig.RequestBufferSize, logger)
	workerPool.Start()
	
	// 创建Protocol Buffers优化器
	protobufOptimizer := grpcOptimizer.NewProtobufOptimizer(grpcOptimizer.DefaultProtobufOptimizerConfig(), logger)

	return &Server{
		config:            cfg,
		logger:            logger,
		grpcServer:        grpcServer,
		handler:           handler,
		connectionPool:    connectionPool,
		workerPool:        workerPool,
		protobufOptimizer: protobufOptimizer,
		authMiddleware:    authMiddleware,
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
	
	// 停止工作池
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
	
	// 关闭连接池
	if s.connectionPool != nil {
		s.connectionPool.Close()
	}
	
	// 优雅停止gRPC服务器
	s.grpcServer.GracefulStop()
	
	s.logger.Info("Server stopped successfully")
	return nil
}

// GetPerformanceStats 获取性能统计信息
func (s *Server) GetPerformanceStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	// 连接池统计
	if s.connectionPool != nil {
		stats["connection_pool"] = s.connectionPool.GetStats()
	}
	
	// 工作池统计
	if s.workerPool != nil {
		stats["worker_pool"] = s.workerPool.GetWorkerPoolStats()
	}
	
	// Protocol Buffers统计
	if s.protobufOptimizer != nil {
		stats["protobuf"] = s.protobufOptimizer.GetCompressionStats()
	}
	
	return stats
}
