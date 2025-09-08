package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cheel98/flashcard-backend/proto/generated/dictionary"
	"github.com/cheel98/flashcard-backend/proto/generated/favorite"
	"github.com/cheel98/flashcard-backend/proto/generated/user"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// ClientManager gRPC客户端连接管理器
type ClientManager struct {
	logger      *zap.Logger
	connections map[string]*grpc.ClientConn
	clients     *Clients
	mu          sync.RWMutex
	config      *ClientConfig
}

// ClientConfig 客户端配置
type ClientConfig struct {
	ServerAddress    string
	MaxConnections   int
	KeepAliveTime    time.Duration
	KeepAliveTimeout time.Duration
	MaxRetries       int
	RetryDelay       time.Duration
}

// Clients 包含所有gRPC客户端
type Clients struct {
	UserService       user.UserServiceClient
	DictionaryService dictionary.DictionaryServiceClient
	FavoriteService   favorite.FavoriteServiceClient
}

// NewClientManager 创建新的客户端管理器
func NewClientManager(logger *zap.Logger, config *ClientConfig) *ClientManager {
	if config == nil {
		config = &ClientConfig{
			ServerAddress:    "localhost:8080",
			MaxConnections:   10,
			KeepAliveTime:    30 * time.Second,
			KeepAliveTimeout: 5 * time.Second,
			MaxRetries:       3,
			RetryDelay:       time.Second,
		}
	}

	return &ClientManager{
		logger:      logger,
		connections: make(map[string]*grpc.ClientConn),
		config:      config,
	}
}

// Connect 建立连接并初始化客户端
func (cm *ClientManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 配置连接选项
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cm.config.KeepAliveTime,
			Timeout:             cm.config.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(4*1024*1024), // 4MB
			grpc.MaxCallSendMsgSize(4*1024*1024), // 4MB
		),
	}

	// 建立连接
	conn, err := grpc.DialContext(ctx, cm.config.ServerAddress, opts...)
	if err != nil {
		cm.logger.Error("Failed to connect to gRPC server",
			zap.String("address", cm.config.ServerAddress),
			zap.Error(err))
		return fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	// 存储连接
	cm.connections["main"] = conn

	// 初始化客户端
	cm.clients = &Clients{
		UserService:       user.NewUserServiceClient(conn),
		DictionaryService: dictionary.NewDictionaryServiceClient(conn),
		FavoriteService:   favorite.NewFavoriteServiceClient(conn),
	}

	cm.logger.Info("gRPC client connected successfully",
		zap.String("address", cm.config.ServerAddress))

	return nil
}

// GetClients 获取gRPC客户端
func (cm *ClientManager) GetClients() *Clients {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.clients
}

// HealthCheck 健康检查
func (cm *ClientManager) HealthCheck(ctx context.Context) error {
	cm.mu.RLock()
	conn, exists := cm.connections["main"]
	cm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no connection available")
	}

	// 检查连接状态
	state := conn.GetState()
	if state.String() != "READY" && state.String() != "IDLE" {
		return fmt.Errorf("connection not ready, state: %s", state.String())
	}

	return nil
}

// Reconnect 重新连接
func (cm *ClientManager) Reconnect(ctx context.Context) error {
	cm.logger.Info("Attempting to reconnect to gRPC server")

	// 关闭现有连接
	cm.Close()

	// 重试连接
	var err error
	for i := 0; i < cm.config.MaxRetries; i++ {
		err = cm.Connect(ctx)
		if err == nil {
			cm.logger.Info("Reconnected to gRPC server successfully")
			return nil
		}

		cm.logger.Warn("Reconnection attempt failed",
			zap.Int("attempt", i+1),
			zap.Int("max_retries", cm.config.MaxRetries),
			zap.Error(err))

		if i < cm.config.MaxRetries-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cm.config.RetryDelay):
				// 继续重试
			}
		}
	}

	return fmt.Errorf("failed to reconnect after %d attempts: %w", cm.config.MaxRetries, err)
}

// Close 关闭所有连接
func (cm *ClientManager) Close() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for name, conn := range cm.connections {
		if err := conn.Close(); err != nil {
			cm.logger.Error("Failed to close connection",
				zap.String("connection", name),
				zap.Error(err))
		} else {
			cm.logger.Info("Connection closed",
				zap.String("connection", name))
		}
	}

	// 清空连接映射
	cm.connections = make(map[string]*grpc.ClientConn)
	cm.clients = nil
}

// GetConnectionCount 获取当前连接数
func (cm *ClientManager) GetConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

// GetConnectionStatus 获取连接状态信息
func (cm *ClientManager) GetConnectionStatus() map[string]string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	status := make(map[string]string)
	for name, conn := range cm.connections {
		status[name] = conn.GetState().String()
	}

	return status
}