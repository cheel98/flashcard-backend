package grpc

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/cheel98/flashcard-backend/proto/generated/health"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// HealthGRPCServer 健康检查gRPC服务实现
type HealthGRPCServer struct {
	health.UnimplementedHealthServiceServer
	logger    *zap.Logger
	services  map[string]health.HealthCheckResponse_ServingStatus
	mu        sync.RWMutex
	startTime time.Time
	stats     *ServerStats
	serverId  string
}

// ServerStats 服务器统计信息
type ServerStats struct {
	mu                sync.RWMutex
	activeConnections int32
	totalRequests     int64
	failedRequests    int64
}

// NewHealthGRPCServer 创建新的健康检查gRPC服务
func NewHealthGRPCServer(logger *zap.Logger) *HealthGRPCServer {
	return &HealthGRPCServer{
		logger:    logger,
		services:  make(map[string]health.HealthCheckResponse_ServingStatus),
		startTime: time.Now(),
		stats:     &ServerStats{},
		serverId:  "flashcard-backend-server",
	}
}

// SetServingStatus 设置服务状态
func (s *HealthGRPCServer) SetServingStatus(service string, status health.HealthCheckResponse_ServingStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.services[service] = status
	s.logger.Info("Service status updated",
		zap.String("service", service),
		zap.String("status", status.String()))
}

// Check 检查服务健康状态
func (s *HealthGRPCServer) Check(ctx context.Context, req *health.HealthCheckRequest) (*health.HealthCheckResponse, error) {
	s.incrementTotalRequests()
	s.logger.Debug("Health check requested", zap.String("service", req.Service))

	s.mu.RLock()
	defer s.mu.RUnlock()

	var status health.HealthCheckResponse_ServingStatus
	var message string

	if req.Service == "" {
		// 检查整个服务器状态
		status = health.HealthCheckResponse_SERVING
		message = "Server is healthy"

		// 如果有任何服务不健康，则整个服务器状态为不健康
		for serviceName, serviceStatus := range s.services {
			if serviceStatus != health.HealthCheckResponse_SERVING {
				status = health.HealthCheckResponse_NOT_SERVING
				message = "Some services are not healthy"
				s.logger.Warn("Unhealthy service detected", zap.String("service", serviceName))
				break
			}
		}
	} else {
		// 检查特定服务状态
		if serviceStatus, exists := s.services[req.Service]; exists {
			status = serviceStatus
			message = "Service status: " + serviceStatus.String()
		} else {
			status = health.HealthCheckResponse_SERVICE_UNKNOWN
			message = "Service not found"
		}
	}

	return &health.HealthCheckResponse{
		Status:    status,
		Message:   message,
		Timestamp: timestamppb.Now(),
	}, nil
}

// Watch 监听服务健康状态变化（流式）
func (s *HealthGRPCServer) Watch(req *health.HealthCheckRequest, stream health.HealthService_WatchServer) error {
	s.logger.Info("Health watch started", zap.String("service", req.Service))
	s.incrementActiveConnections()
	defer s.decrementActiveConnections()

	// 首先发送当前状态
	currentStatus, err := s.Check(stream.Context(), req)
	if err != nil {
		return err
	}

	if err := stream.Send(currentStatus); err != nil {
		s.logger.Error("Failed to send initial health status", zap.Error(err))
		return err
	}

	// 定期发送状态更新
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("Health watch stream closed", zap.String("service", req.Service))
			return stream.Context().Err()
		case <-ticker.C:
			status, err := s.Check(stream.Context(), req)
			if err != nil {
				s.logger.Error("Failed to check health status", zap.Error(err))
				return err
			}
			if err := stream.Send(status); err != nil {
				s.logger.Error("Failed to send health status update", zap.Error(err))
				return err
			}
		}
	}
}

// Heartbeat 心跳检查
func (s *HealthGRPCServer) Heartbeat(ctx context.Context, req *health.HeartbeatRequest) (*health.HeartbeatResponse, error) {
	s.incrementTotalRequests()
	s.logger.Debug("Heartbeat received", zap.String("client_id", req.ClientId))

	// 获取服务器统计信息
	stats := s.getServerStats()

	return &health.HeartbeatResponse{
		ServerId:  s.serverId,
		Timestamp: timestamppb.Now(),
		Alive:     true,
		Stats:     stats,
	}, nil
}

// getServerStats 获取服务器统计信息
func (s *HealthGRPCServer) getServerStats() *health.ServerStats {
	s.stats.mu.RLock()
	defer s.stats.mu.RUnlock()

	// 获取系统内存信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 计算运行时间
	uptime := time.Since(s.startTime)

	return &health.ServerStats{
		UptimeSeconds:     int64(uptime.Seconds()),
		ActiveConnections: s.stats.activeConnections,
		CpuUsage:          0.0,                               // 简化实现，实际项目中可以使用第三方库获取CPU使用率
		MemoryUsage:       float64(m.Alloc) / float64(m.Sys), // 内存使用率
		TotalRequests:     s.stats.totalRequests,
		FailedRequests:    s.stats.failedRequests,
	}
}

// incrementActiveConnections 增加活跃连接数
func (s *HealthGRPCServer) incrementActiveConnections() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()
	s.stats.activeConnections++
}

// decrementActiveConnections 减少活跃连接数
func (s *HealthGRPCServer) decrementActiveConnections() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()
	s.stats.activeConnections--
}

// incrementTotalRequests 增加总请求数
func (s *HealthGRPCServer) incrementTotalRequests() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()
	s.stats.totalRequests++
}

// incrementFailedRequests 增加失败请求数
func (s *HealthGRPCServer) incrementFailedRequests() {
	s.stats.mu.Lock()
	defer s.stats.mu.Unlock()
	s.stats.failedRequests++
}

// InitializeServices 初始化服务状态
func (s *HealthGRPCServer) InitializeServices() {
	// 设置默认服务状态
	s.SetServingStatus("UserService", health.HealthCheckResponse_SERVING)
	s.SetServingStatus("DictionaryService", health.HealthCheckResponse_SERVING)
	s.SetServingStatus("FavoriteService", health.HealthCheckResponse_SERVING)
	s.SetServingStatus("HealthService", health.HealthCheckResponse_SERVING)

	s.logger.Info("Health service initialized with default service statuses")
}
