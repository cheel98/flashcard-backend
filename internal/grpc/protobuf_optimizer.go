package grpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ProtobufOptimizerConfig Protocol Buffers优化配置
type ProtobufOptimizerConfig struct {
	EnableCompression     bool    `json:"enable_compression"`
	CompressionThreshold  int     `json:"compression_threshold"`  // 字节数阈值
	CompressionLevel      int     `json:"compression_level"`      // gzip压缩级别
	EnablePooling         bool    `json:"enable_pooling"`         // 启用对象池
	PoolMaxSize           int     `json:"pool_max_size"`          // 对象池最大大小
	EnableMetrics         bool    `json:"enable_metrics"`         // 启用性能指标
	MaxMessageSize        int     `json:"max_message_size"`       // 最大消息大小
	SerializationTimeout  time.Duration `json:"serialization_timeout"`  // 序列化超时
}

// DefaultProtobufOptimizerConfig 返回默认配置
func DefaultProtobufOptimizerConfig() *ProtobufOptimizerConfig {
	return &ProtobufOptimizerConfig{
		EnableCompression:    true,
		CompressionThreshold: 1024, // 1KB
		CompressionLevel:     gzip.DefaultCompression,
		EnablePooling:        true,
		PoolMaxSize:          100,
		EnableMetrics:        true,
		MaxMessageSize:       4 * 1024 * 1024, // 4MB
		SerializationTimeout: 5 * time.Second,
	}
}

// ProtobufOptimizer Protocol Buffers优化器
type ProtobufOptimizer struct {
	config      *ProtobufOptimizerConfig
	bufferPool  *sync.Pool
	compressors *sync.Pool
	metrics     *ProtobufMetrics
	logger      *zap.Logger
}

// ProtobufMetrics 性能指标
type ProtobufMetrics struct {
	SerializationCount   int64         `json:"serialization_count"`
	DeserializationCount int64         `json:"deserialization_count"`
	CompressionCount     int64         `json:"compression_count"`
	TotalSerializedSize  int64         `json:"total_serialized_size"`
	TotalCompressedSize  int64         `json:"total_compressed_size"`
	AvgSerializationTime time.Duration `json:"avg_serialization_time"`
	AvgCompressionRatio  float64       `json:"avg_compression_ratio"`
	mutex                sync.RWMutex
}

// NewProtobufOptimizer 创建新的Protocol Buffers优化器
func NewProtobufOptimizer(config *ProtobufOptimizerConfig, logger *zap.Logger) *ProtobufOptimizer {
	if config == nil {
		config = DefaultProtobufOptimizerConfig()
	}

	optimizer := &ProtobufOptimizer{
		config: config,
		logger: logger,
	}

	// 初始化缓冲区池
	if config.EnablePooling {
		optimizer.bufferPool = &sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, 1024))
			},
		}

		// 初始化压缩器池
		optimizer.compressors = &sync.Pool{
			New: func() interface{} {
				var buf bytes.Buffer
				writer, _ := gzip.NewWriterLevel(&buf, config.CompressionLevel)
				return writer
			},
		}
	}

	// 初始化性能指标
	if config.EnableMetrics {
		optimizer.metrics = &ProtobufMetrics{}
	}

	return optimizer
}

// SerializeMessage 序列化消息
func (po *ProtobufOptimizer) SerializeMessage(ctx context.Context, msg proto.Message) ([]byte, error) {
	start := time.Now()
	defer func() {
		if po.config.EnableMetrics {
			po.updateSerializationMetrics(time.Since(start))
		}
	}()

	// 检查上下文超时
	if po.config.SerializationTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, po.config.SerializationTimeout)
		defer cancel()
	}

	// 序列化消息
	data, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal protobuf message: %w", err)
	}

	// 检查消息大小
	if len(data) > po.config.MaxMessageSize {
		return nil, fmt.Errorf("message size %d exceeds maximum %d", len(data), po.config.MaxMessageSize)
	}

	// 检查是否需要压缩
	if po.config.EnableCompression && len(data) > po.config.CompressionThreshold {
		compressedData, err := po.compressData(data)
		if err != nil {
			po.logger.Warn("Failed to compress data, using uncompressed", zap.Error(err))
			return data, nil
		}

		// 如果压缩后更小，使用压缩数据
		if len(compressedData) < len(data) {
			if po.config.EnableMetrics {
				po.updateCompressionMetrics(len(data), len(compressedData))
			}
			return compressedData, nil
		}
	}

	return data, nil
}

// DeserializeMessage 反序列化消息
func (po *ProtobufOptimizer) DeserializeMessage(ctx context.Context, data []byte, msg proto.Message) error {
	start := time.Now()
	defer func() {
		if po.config.EnableMetrics {
			po.updateDeserializationMetrics(time.Since(start))
		}
	}()

	// 检查上下文超时
	if po.config.SerializationTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, po.config.SerializationTimeout)
		defer cancel()
	}

	// 检查数据大小
	if len(data) > po.config.MaxMessageSize {
		return fmt.Errorf("data size %d exceeds maximum %d", len(data), po.config.MaxMessageSize)
	}

	// 尝试检测是否为压缩数据
	if po.config.EnableCompression && po.isCompressedData(data) {
		decompressedData, err := po.decompressData(data)
		if err != nil {
			po.logger.Warn("Failed to decompress data, trying direct unmarshal", zap.Error(err))
		} else {
			data = decompressedData
		}
	}

	// 反序列化消息
	return proto.Unmarshal(data, msg)
}

// compressData 压缩数据
func (po *ProtobufOptimizer) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	var writer *gzip.Writer

	// 使用对象池获取压缩器
	if po.config.EnablePooling && po.compressors != nil {
		writer = po.compressors.Get().(*gzip.Writer)
		writer.Reset(&buf)
		defer po.compressors.Put(writer)
	} else {
		var err error
		writer, err = gzip.NewWriterLevel(&buf, po.config.CompressionLevel)
		if err != nil {
			return nil, err
		}
		defer writer.Close()
	}

	// 写入数据
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	// 关闭压缩器
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData 解压缩数据
func (po *ProtobufOptimizer) decompressData(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	if po.config.EnablePooling && po.bufferPool != nil {
		bufPtr := po.bufferPool.Get().(*bytes.Buffer)
		bufPtr.Reset()
		defer po.bufferPool.Put(bufPtr)
		buf = *bufPtr
	}

	// 限制解压缩大小
	limitedReader := io.LimitReader(reader, int64(po.config.MaxMessageSize))
	if _, err := io.Copy(&buf, limitedReader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// isCompressedData 检测数据是否为gzip压缩格式
func (po *ProtobufOptimizer) isCompressedData(data []byte) bool {
	// gzip魔数检测
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// updateSerializationMetrics 更新序列化指标
func (po *ProtobufOptimizer) updateSerializationMetrics(duration time.Duration) {
	if po.metrics == nil {
		return
	}

	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.SerializationCount++
	// 计算平均序列化时间
	if po.metrics.SerializationCount == 1 {
		po.metrics.AvgSerializationTime = duration
	} else {
		// 使用指数移动平均
		alpha := 0.1
		po.metrics.AvgSerializationTime = time.Duration(float64(po.metrics.AvgSerializationTime)*(1-alpha) + float64(duration)*alpha)
	}
}

// updateDeserializationMetrics 更新反序列化指标
func (po *ProtobufOptimizer) updateDeserializationMetrics(duration time.Duration) {
	if po.metrics == nil {
		return
	}

	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.DeserializationCount++
}

// updateCompressionMetrics 更新压缩指标
func (po *ProtobufOptimizer) updateCompressionMetrics(originalSize, compressedSize int) {
	if po.metrics == nil {
		return
	}

	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.CompressionCount++
	po.metrics.TotalSerializedSize += int64(originalSize)
	po.metrics.TotalCompressedSize += int64(compressedSize)

	// 计算平均压缩比
	if po.metrics.TotalSerializedSize > 0 {
		po.metrics.AvgCompressionRatio = float64(po.metrics.TotalCompressedSize) / float64(po.metrics.TotalSerializedSize)
	}
}

// GetMetrics 获取性能指标
func (po *ProtobufOptimizer) GetMetrics() *ProtobufMetrics {
	if po.metrics == nil {
		return nil
	}

	po.metrics.mutex.RLock()
	defer po.metrics.mutex.RUnlock()

	// 返回指标副本
	return &ProtobufMetrics{
		SerializationCount:   po.metrics.SerializationCount,
		DeserializationCount: po.metrics.DeserializationCount,
		CompressionCount:     po.metrics.CompressionCount,
		TotalSerializedSize:  po.metrics.TotalSerializedSize,
		TotalCompressedSize:  po.metrics.TotalCompressedSize,
		AvgSerializationTime: po.metrics.AvgSerializationTime,
		AvgCompressionRatio:  po.metrics.AvgCompressionRatio,
	}
}

// ResetMetrics 重置性能指标
func (po *ProtobufOptimizer) ResetMetrics() {
	if po.metrics == nil {
		return
	}

	po.metrics.mutex.Lock()
	defer po.metrics.mutex.Unlock()

	po.metrics.SerializationCount = 0
	po.metrics.DeserializationCount = 0
	po.metrics.CompressionCount = 0
	po.metrics.TotalSerializedSize = 0
	po.metrics.TotalCompressedSize = 0
	po.metrics.AvgSerializationTime = 0
	po.metrics.AvgCompressionRatio = 0
}

// ValidateMessage 验证消息
func (po *ProtobufOptimizer) ValidateMessage(msg proto.Message) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}

	// 检查消息是否有效
	if !msg.ProtoReflect().IsValid() {
		return fmt.Errorf("message is not valid")
	}

	// 检查必填字段
	reflectMsg := msg.ProtoReflect()
	descriptor := reflectMsg.Descriptor()
	fields := descriptor.Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		if field.Cardinality() == protoreflect.Required {
			if !reflectMsg.Has(field) {
				return fmt.Errorf("required field %s is missing", field.Name())
			}
		}
	}

	return nil
}

// GetCompressionStats 获取压缩统计信息
func (po *ProtobufOptimizer) GetCompressionStats() map[string]interface{} {
	metrics := po.GetMetrics()
	if metrics == nil {
		return nil
	}

	stats := map[string]interface{}{
		"compression_enabled":     po.config.EnableCompression,
		"compression_threshold":   po.config.CompressionThreshold,
		"compression_level":       po.config.CompressionLevel,
		"compression_count":       metrics.CompressionCount,
		"total_serialized_size":   metrics.TotalSerializedSize,
		"total_compressed_size":   metrics.TotalCompressedSize,
		"avg_compression_ratio":   metrics.AvgCompressionRatio,
		"serialization_count":     metrics.SerializationCount,
		"deserialization_count":   metrics.DeserializationCount,
		"avg_serialization_time":  metrics.AvgSerializationTime.String(),
	}

	if metrics.TotalSerializedSize > 0 {
		stats["space_saved_bytes"] = metrics.TotalSerializedSize - metrics.TotalCompressedSize
		stats["space_saved_percentage"] = (1 - metrics.AvgCompressionRatio) * 100
	}

	return stats
}