package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cheel98/flashcard-backend/internal/config"
	grpcOptimizer "github.com/cheel98/flashcard-backend/int
	"github.com/cheel98/flashcard-backend/internal/handler"
	"github.com/cheel98/flashcard-backend/internal/middleware"

	// gRPC-Gateway 相关导入
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	// 生成的 gRPC-Gateway 代码
	dictionaryPb "github.com/cheel98/flashcard-backend/proto/generated/dictionary"
	favoritePb "github.com/cheel98/flashcard-backend/proto/generated/favorite"
	healthPb "github.com/cheel98/flashcard-backend/proto/generated/health"
	translationPb "github.com/cheel98/flashcard-backend/proto/generated/translation"
	userPb "github.com/cheel98/flashcard-backend/proto/generated/user"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server 服务器结构体
type Server struct {
	config            *config.Config
	logger            *zap.Logger
	grpcServer        *grpc.Server
	httpServer        *http.Server
	handler           *handler.Handler
	connectionPool    *grpcOptimizer.ConnectionPool
	workerPool        *grpcOptimizer.WorkerPool
	protobufOptimizer *grpcOptimizer.ProtobufOptimizer
	authMiddleware    *middleware.AuthMiddleware
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

	// 监听gRPC端口
	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Server.Port))
	if err != nil {
		s.logger.Error("Failed to listen gRPC port", zap.Error(err))
		return err
	}

	// 启动gRPC服务器
	go func() {
		s.logger.Info("gRPC Server starting", zap.Int("port", s.config.Server.Port))
		if err := s.grpcServer.Serve(grpcLis); err != nil {
			s.logger.Error("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// 启动gRPC-Gateway HTTP服务器
	go func() {
		if err := s.startHTTPGateway(); err != nil {
			s.logger.Error("Failed to start HTTP gateway", zap.Error(err))
		}
	}()

	return nil
}

// startHTTPGateway 启动gRPC-Gateway HTTP服务器
func (s *Server) startHTTPGateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// 创建gRPC-Gateway mux
	mux := runtime.NewServeMux()

	// gRPC服务器地址
	grpcServerEndpoint := fmt.Sprintf("localhost:%d", s.config.Server.Port)

	// 连接选项
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// 注册所有服务的HTTP处理器
	if err := userPb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register user service handler: %w", err)
	}

	if err := dictionaryPb.RegisterDictionaryServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register dictionary service handler: %w", err)
	}

	if err := favoritePb.RegisterFavoriteServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register favorite service handler: %w", err)
	}

	if err := translationPb.RegisterTranslationHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register translation service handler: %w", err)
	}

	if err := healthPb.RegisterHealthServiceHandlerFromEndpoint(ctx, mux, grpcServerEndpoint, opts); err != nil {
		return fmt.Errorf("failed to register health service handler: %w", err)
	}

	// 创建HTTP服务器
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Server.HTTPPort),
		Handler: mux,
	}

	s.logger.Info("HTTP Gateway starting", zap.Int("port", s.config.Server.HTTPPort))

	// 启动HTTP服务器
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve HTTP gateway: %w", err)
	}

	return nil
}

// Stop 停止服务器
func (s *Server) Stop() error {
	s.logger.Info("Stopping server...")

	// 停止HTTP服务器
	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		}
	}

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
