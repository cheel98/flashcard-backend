package handler

import (
	"github.com/cheel98/flashcard-backend/internal/grpc"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/pkg/email"
	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"github.com/cheel98/flashcard-backend/pkg/redis"
	"github.com/cheel98/flashcard-backend/proto/generated/dictionary"
	"github.com/cheel98/flashcard-backend/proto/generated/favorite"
	"github.com/cheel98/flashcard-backend/proto/generated/health"
	"github.com/cheel98/flashcard-backend/proto/generated/user"
	"go.uber.org/zap"
	grpcServer "google.golang.org/grpc"
)

// Handler gRPC处理器
type Handler struct {
	logger         *zap.Logger
	userRepo       repository.UserRepository
	dictionaryRepo repository.DictionaryRepository
	favoriteRepo   repository.FavoriteRepository
	jwtManager     *jwt.JWTManager
	EmailService   *email.EmailService
	RedisClient    *redis.RedisClient
	healthServer   *grpc.HealthGRPCServer
}

// NewHandler 创建新的处理器
func NewHandler(
	logger *zap.Logger,
	userRepo repository.UserRepository,
	dictionaryRepo repository.DictionaryRepository,
	favoriteRepo repository.FavoriteRepository,
	jwtManager *jwt.JWTManager,
	email *email.EmailService,
	redisClient *redis.RedisClient,
) *Handler {
	// 创建健康检查服务
	healthServer := grpc.NewHealthGRPCServer(logger)
	healthServer.InitializeServices()

	return &Handler{
		logger:         logger,
		userRepo:       userRepo,
		dictionaryRepo: dictionaryRepo,
		favoriteRepo:   favoriteRepo,
		jwtManager:     jwtManager,
		EmailService:   email,
		RedisClient:    redisClient,
		healthServer:   healthServer,
	}
}

// RegisterServices 注册gRPC服务
func (h *Handler) RegisterServices(server *grpcServer.Server) {
	// 创建gRPC服务实例
	userGRPCServer := grpc.NewUserGRPCServer(h.userRepo, h.jwtManager, h.RedisClient, h.EmailService, h.logger)
	dictionaryGRPCServer := grpc.NewDictionaryGRPCServer(h.dictionaryRepo, h.logger)
	favoriteGRPCServer := grpc.NewFavoriteGRPCServer(h.favoriteRepo, h.logger)

	// 注册gRPC服务
	user.RegisterUserServiceServer(server, userGRPCServer)
	dictionary.RegisterDictionaryServiceServer(server, dictionaryGRPCServer)
	favorite.RegisterFavoriteServiceServer(server, favoriteGRPCServer)
	health.RegisterHealthServiceServer(server, h.healthServer)

	h.logger.Info("gRPC services registered successfully",
		zap.String("services", "UserService, DictionaryService, FavoriteService, HealthService"))
}

// GetHealthServer 获取健康检查服务实例
func (h *Handler) GetHealthServer() *grpc.HealthGRPCServer {
	return h.healthServer
}
