package handler

import (
	"github.com/cheel98/flashcard-backend/internal/grpc"
	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/cheel98/flashcard-backend/proto/generated/dictionary"
	"github.com/cheel98/flashcard-backend/proto/generated/favorite"
	"github.com/cheel98/flashcard-backend/proto/generated/user"
	"go.uber.org/zap"
	grpcServer "google.golang.org/grpc"
)

// Handler gRPC处理器
type Handler struct {
	logger            *zap.Logger
	userService       service.UserService
	dictionaryService service.DictionaryService
	favoriteService   service.FavoriteService
}

// NewHandler 创建新的处理器
func NewHandler(
	logger *zap.Logger,
	userService service.UserService,
	dictionaryService service.DictionaryService,
	favoriteService service.FavoriteService,
) *Handler {
	return &Handler{
		logger:            logger,
		userService:       userService,
		dictionaryService: dictionaryService,
		favoriteService:   favoriteService,
	}
}

// RegisterServices 注册gRPC服务
func (h *Handler) RegisterServices(server *grpcServer.Server) {
	// 创建gRPC服务实例
	userGRPCServer := grpc.NewUserGRPCServer(h.userService, h.logger)
	dictionaryGRPCServer := grpc.NewDictionaryGRPCServer(h.dictionaryService, h.logger)
	favoriteGRPCServer := grpc.NewFavoriteGRPCServer(h.favoriteService, h.logger)

	// 注册gRPC服务
	user.RegisterUserServiceServer(server, userGRPCServer)
	dictionary.RegisterDictionaryServiceServer(server, dictionaryGRPCServer)
	favorite.RegisterFavoriteServiceServer(server, favoriteGRPCServer)

	h.logger.Info("gRPC services registered successfully",
		zap.String("services", "UserService, DictionaryService, FavoriteService"))
}
