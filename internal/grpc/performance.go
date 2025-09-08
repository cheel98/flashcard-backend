package grpc

import (
	"context"
	"runtime"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// PerformanceConfig gRPC性能优化配置
type PerformanceConfig struct {
	// 服务器配置
	MaxConcurrentStreams  uint32        `json:"max_concurrent_streams"`
	MaxReceiveMessageSize int           `json:"max_receive_message_size"`
	MaxSendMessageSize    int           `json:"max_send_message_size"`
	ConnectionTimeout     time.Duration `json:"connection_timeout"`
	
	// Keep-Alive配置
	KeepAliveTime             time.Duration `json:"keep_alive_time"`
	KeepAliveTimeout          time.Duration `json:"keep_alive_timeout"`
	KeepAliveEnforcementMinTime time.Duration `json:"keep_alive_enforcement_min_time"`
	KeepAliveEnforcementPermitWithoutStream bool `json:"keep_alive_enforcement_permit_without_stream"`
	
	// 并发配置
	WorkerPoolSize    int `json:"worker_pool_size"`
	MaxWorkers        int `json:"max_workers"`
	RequestBufferSize int `json:"request_buffer_size"`
	
	// 压缩配置
	EnableCompression bool   `json:"enable_compression"`
	CompressionLevel  string `json:"compression_level"`
}

// DefaultPerformanceConfig 返回默认性能配置
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		// 服务器配置
		MaxConcurrentStreams:  1000,
		MaxReceiveMessageSize: 4 * 1024 * 1024, // 4MB
		MaxSendMessageSize:    4 * 1024 * 1024, // 4MB
		ConnectionTimeout:     30 * time.Second,
		
		// Keep-Alive配置
		KeepAliveTime:             30 * time.Second,
		KeepAliveTimeout:          5 * time.Second,
		KeepAliveEnforcementMinTime: 5 * time.Second,
		KeepAliveEnforcementPermitWithoutStream: true,
		
		// 并发配置
		WorkerPoolSize:    runtime.NumCPU() * 2,
		MaxWorkers:        runtime.NumCPU() * 4,
		RequestBufferSize: 1000,
		
		// 压缩配置
		EnableCompression: true,
		CompressionLevel:  "gzip",
	}
}

// CreateOptimizedServer 创建性能优化的gRPC服务器
func CreateOptimizedServer(config *PerformanceConfig, logger *zap.Logger) *grpc.Server {
	return CreateOptimizedServerWithInterceptors(config, logger, nil, nil)
}

// CreateOptimizedServerWithInterceptors 创建带有自定义拦截器的性能优化gRPC服务器
func CreateOptimizedServerWithInterceptors(
	config *PerformanceConfig, 
	logger *zap.Logger,
	unaryInterceptor grpc.UnaryServerInterceptor,
	streamInterceptor grpc.StreamServerInterceptor,
) *grpc.Server {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	// 服务器选项
	opts := []grpc.ServerOption{
		// 并发流限制
		grpc.MaxConcurrentStreams(config.MaxConcurrentStreams),
		
		// 消息大小限制
		grpc.MaxRecvMsgSize(config.MaxReceiveMessageSize),
		grpc.MaxSendMsgSize(config.MaxSendMessageSize),
		
		// Keep-Alive配置
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    config.KeepAliveTime,
			Timeout: config.KeepAliveTimeout,
		}),
		
		// Keep-Alive执行策略
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.KeepAliveEnforcementMinTime,
			PermitWithoutStream: config.KeepAliveEnforcementPermitWithoutStream,
		}),
		
		// 连接超时
		grpc.ConnectionTimeout(config.ConnectionTimeout),
	}

	// 创建拦截器链
	var unaryInterceptors []grpc.UnaryServerInterceptor
	var streamInterceptors []grpc.StreamServerInterceptor

	// 添加性能监控拦截器
	unaryInterceptors = append(unaryInterceptors, performanceUnaryInterceptor(logger))
	streamInterceptors = append(streamInterceptors, performanceStreamInterceptor(logger))

	// 添加自定义拦截器（如JWT认证）
	if unaryInterceptor != nil {
		unaryInterceptors = append(unaryInterceptors, unaryInterceptor)
	}
	if streamInterceptor != nil {
		streamInterceptors = append(streamInterceptors, streamInterceptor)
	}

	// 添加拦截器链到服务器选项
	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.ChainStreamInterceptor(streamInterceptors...))
	}

	// 创建服务器
	server := grpc.NewServer(opts...)

	// 启用反射（开发环境）
	reflection.Register(server)

	logger.Info("Optimized gRPC server created",
		zap.Uint32("max_concurrent_streams", config.MaxConcurrentStreams),
		zap.Int("max_recv_msg_size", config.MaxReceiveMessageSize),
		zap.Int("max_send_msg_size", config.MaxSendMessageSize),
		zap.Duration("keep_alive_time", config.KeepAliveTime),
		zap.Duration("keep_alive_timeout", config.KeepAliveTimeout))

	return server
}

// performanceUnaryInterceptor 一元RPC性能拦截器
func performanceUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		
		// 执行RPC调用
		resp, err := handler(ctx, req)
		
		// 记录性能指标
		duration := time.Since(start)
		logger.Debug("Unary RPC completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Bool("success", err == nil))
		
		// 如果请求时间过长，记录警告
		if duration > 5*time.Second {
			logger.Warn("Slow RPC detected",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", duration))
		}
		
		return resp, err
	}
}

// performanceStreamInterceptor 流式RPC性能拦截器
func performanceStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		
		// 执行流式RPC调用
		err := handler(srv, stream)
		
		// 记录性能指标
		duration := time.Since(start)
		logger.Debug("Stream RPC completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Bool("success", err == nil))
		
		return err
	}
}

// WorkerPool 工作池结构
type WorkerPool struct {
	workerCount int
	jobQueue    chan func()
	quit        chan bool
	logger      *zap.Logger
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(workerCount int, queueSize int, logger *zap.Logger) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan func(), queueSize),
		quit:        make(chan bool),
		logger:      logger,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(i)
	}
	wp.logger.Info("Worker pool started", zap.Int("workers", wp.workerCount))
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	wp.logger.Info("Worker pool stopped")
}

// Submit 提交任务到工作池
func (wp *WorkerPool) Submit(job func()) {
	select {
	case wp.jobQueue <- job:
		// 任务已提交
	default:
		// 队列已满，直接执行
		wp.logger.Warn("Worker pool queue full, executing job directly")
		go job()
	}
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	for {
		select {
		case job := <-wp.jobQueue:
			// 执行任务
			job()
		case <-wp.quit:
			// 退出工作协程
			wp.logger.Debug("Worker stopped", zap.Int("worker_id", id))
			return
		}
	}
}

// GetWorkerPoolStats 获取工作池统计信息
func (wp *WorkerPool) GetWorkerPoolStats() map[string]interface{} {
	return map[string]interface{}{
		"worker_count":    wp.workerCount,
		"queue_length":    len(wp.jobQueue),
		"queue_capacity":  cap(wp.jobQueue),
		"queue_usage":     float64(len(wp.jobQueue)) / float64(cap(wp.jobQueue)) * 100,
	}
}