package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/keepalive"
)

// ConnectionPoolConfig 连接池配置
type ConnectionPoolConfig struct {
	MaxConnections    int           `json:"max_connections"`
	MinConnections    int           `json:"min_connections"`
	MaxIdleTime       time.Duration `json:"max_idle_time"`
	HealthCheckPeriod time.Duration `json:"health_check_period"`
	ConnectTimeout    time.Duration `json:"connect_timeout"`
	KeepAliveTime     time.Duration `json:"keep_alive_time"`
	KeepAliveTimeout  time.Duration `json:"keep_alive_timeout"`
}

// DefaultConnectionPoolConfig 返回默认连接池配置
func DefaultConnectionPoolConfig() *ConnectionPoolConfig {
	return &ConnectionPoolConfig{
		MaxConnections:    10,
		MinConnections:    2,
		MaxIdleTime:       5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
		ConnectTimeout:    10 * time.Second,
		KeepAliveTime:     30 * time.Second,
		KeepAliveTimeout:  5 * time.Second,
	}
}

// PooledConnection 池化连接
type PooledConnection struct {
	conn        *grpc.ClientConn
	lastUsed    time.Time
	inUse       bool
	createdAt   time.Time
	useCount    int64
	mutex       sync.RWMutex
}

// IsHealthy 检查连接是否健康
func (pc *PooledConnection) IsHealthy() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	
	if pc.conn == nil {
		return false
	}
	
	state := pc.conn.GetState()
	return state == connectivity.Ready || state == connectivity.Idle
}

// MarkUsed 标记连接为使用中
func (pc *PooledConnection) MarkUsed() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	pc.inUse = true
	pc.lastUsed = time.Now()
	pc.useCount++
}

// MarkIdle 标记连接为空闲
func (pc *PooledConnection) MarkIdle() {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	pc.inUse = false
	pc.lastUsed = time.Now()
}

// IsIdle 检查连接是否空闲
func (pc *PooledConnection) IsIdle() bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return !pc.inUse
}

// IsExpired 检查连接是否过期
func (pc *PooledConnection) IsExpired(maxIdleTime time.Duration) bool {
	pc.mutex.RLock()
	defer pc.mutex.RUnlock()
	return time.Since(pc.lastUsed) > maxIdleTime
}

// Close 关闭连接
func (pc *PooledConnection) Close() error {
	pc.mutex.Lock()
	defer pc.mutex.Unlock()
	
	if pc.conn != nil {
		return pc.conn.Close()
	}
	return nil
}

// ConnectionPool gRPC连接池
type ConnectionPool struct {
	config      *ConnectionPoolConfig
	connections map[string][]*PooledConnection
	mutex       sync.RWMutex
	logger      *zap.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewConnectionPool 创建新的连接池
func NewConnectionPool(config *ConnectionPoolConfig, logger *zap.Logger) *ConnectionPool {
	if config == nil {
		config = DefaultConnectionPoolConfig()
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &ConnectionPool{
		config:      config,
		connections: make(map[string][]*PooledConnection),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// 启动健康检查和清理协程
	pool.wg.Add(1)
	go pool.healthCheckLoop()
	
	return pool
}

// GetConnection 获取连接
func (cp *ConnectionPool) GetConnection(target string) (*grpc.ClientConn, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	
	// 查找可用的空闲连接
	connections := cp.connections[target]
	for _, pooledConn := range connections {
		if pooledConn.IsIdle() && pooledConn.IsHealthy() {
			pooledConn.MarkUsed()
			cp.logger.Debug("Reusing existing connection", zap.String("target", target))
			return pooledConn.conn, nil
		}
	}
	
	// 如果没有可用连接且未达到最大连接数，创建新连接
	if len(connections) < cp.config.MaxConnections {
		conn, err := cp.createConnection(target)
		if err != nil {
			return nil, fmt.Errorf("failed to create connection to %s: %w", target, err)
		}
		
		pooledConn := &PooledConnection{
			conn:      conn,
			lastUsed:  time.Now(),
			inUse:     true,
			createdAt: time.Now(),
			useCount:  1,
		}
		
		cp.connections[target] = append(cp.connections[target], pooledConn)
		cp.logger.Info("Created new connection", 
			zap.String("target", target),
			zap.Int("total_connections", len(cp.connections[target])))
		
		return conn, nil
	}
	
	return nil, fmt.Errorf("connection pool exhausted for target %s", target)
}

// ReleaseConnection 释放连接
func (cp *ConnectionPool) ReleaseConnection(target string, conn *grpc.ClientConn) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	
	connections := cp.connections[target]
	for _, pooledConn := range connections {
		if pooledConn.conn == conn {
			pooledConn.MarkIdle()
			cp.logger.Debug("Released connection", zap.String("target", target))
			return
		}
	}
}

// createConnection 创建新的gRPC连接
func (cp *ConnectionPool) createConnection(target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(cp.ctx, cp.config.ConnectTimeout)
	defer cancel()
	
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                cp.config.KeepAliveTime,
			Timeout:             cp.config.KeepAliveTimeout,
			PermitWithoutStream: true,
		}),
	}
	
	return grpc.DialContext(ctx, target, opts...)
}

// healthCheckLoop 健康检查循环
func (cp *ConnectionPool) healthCheckLoop() {
	defer cp.wg.Done()
	
	ticker := time.NewTicker(cp.config.HealthCheckPeriod)
	defer ticker.Stop()
	
	for {
		select {
		case <-cp.ctx.Done():
			return
		case <-ticker.C:
			cp.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (cp *ConnectionPool) performHealthCheck() {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	
	for target, connections := range cp.connections {
		var healthyConnections []*PooledConnection
		
		for _, pooledConn := range connections {
			// 检查连接健康状态
			if !pooledConn.IsHealthy() {
				cp.logger.Warn("Removing unhealthy connection", zap.String("target", target))
				pooledConn.Close()
				continue
			}
			
			// 检查连接是否过期
			if pooledConn.IsIdle() && pooledConn.IsExpired(cp.config.MaxIdleTime) {
				cp.logger.Debug("Removing expired connection", zap.String("target", target))
				pooledConn.Close()
				continue
			}
			
			healthyConnections = append(healthyConnections, pooledConn)
		}
		
		cp.connections[target] = healthyConnections
		
		// 确保最小连接数
		if len(healthyConnections) < cp.config.MinConnections {
			needed := cp.config.MinConnections - len(healthyConnections)
			for i := 0; i < needed; i++ {
				conn, err := cp.createConnection(target)
				if err != nil {
					cp.logger.Error("Failed to create minimum connection", 
						zap.String("target", target), 
						zap.Error(err))
					continue
				}
				
				pooledConn := &PooledConnection{
					conn:      conn,
					lastUsed:  time.Now(),
					inUse:     false,
					createdAt: time.Now(),
					useCount:  0,
				}
				
				cp.connections[target] = append(cp.connections[target], pooledConn)
			}
		}
	}
}

// GetStats 获取连接池统计信息
func (cp *ConnectionPool) GetStats() map[string]interface{} {
	cp.mutex.RLock()
	defer cp.mutex.RUnlock()
	
	stats := make(map[string]interface{})
	totalConnections := 0
	totalActiveConnections := 0
	
	for target, connections := range cp.connections {
		activeCount := 0
		for _, conn := range connections {
			if !conn.IsIdle() {
				activeCount++
			}
		}
		
		stats[target] = map[string]interface{}{
			"total":  len(connections),
			"active": activeCount,
			"idle":   len(connections) - activeCount,
		}
		
		totalConnections += len(connections)
		totalActiveConnections += activeCount
	}
	
	stats["summary"] = map[string]interface{}{
		"total_connections":  totalConnections,
		"active_connections": totalActiveConnections,
		"idle_connections":   totalConnections - totalActiveConnections,
		"targets":            len(cp.connections),
	}
	
	return stats
}

// Close 关闭连接池
func (cp *ConnectionPool) Close() error {
	cp.cancel()
	cp.wg.Wait()
	
	cp.mutex.Lock()
	defer cp.mutex.Unlock()
	
	for target, connections := range cp.connections {
		for _, pooledConn := range connections {
			if err := pooledConn.Close(); err != nil {
				cp.logger.Error("Failed to close connection", 
					zap.String("target", target), 
					zap.Error(err))
			}
		}
	}
	
	cp.connections = make(map[string][]*PooledConnection)
	cp.logger.Info("Connection pool closed")
	
	return nil
}