package grpc

import (
	"context"
	"time"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/proto/generated/favorite"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FavoriteGRPCServer gRPC收藏服务实现
type FavoriteGRPCServer struct {
	favorite.UnimplementedFavoriteServiceServer
	favoriteRepo repository.FavoriteRepository
	logger       *zap.Logger
}

// NewFavoriteGRPCServer 创建新的gRPC收藏服务
func NewFavoriteGRPCServer(favoriteRepo repository.FavoriteRepository, logger *zap.Logger) *FavoriteGRPCServer {
	return &FavoriteGRPCServer{
		favoriteRepo: favoriteRepo,
		logger:       logger,
	}
}

// AddFavorite 添加收藏
func (s *FavoriteGRPCServer) AddFavorite(ctx context.Context, req *favorite.AddFavoriteRequest) (*favorite.AddFavoriteResponse, error) {
	s.logger.Info("添加收藏",
		zap.String("userID", req.UserId),
		zap.Uint64("dictionaryID", req.DictionaryId))

	// 验证必填字段
	if req.UserId == "" || req.DictionaryId == 0 {
		s.logger.Error("添加收藏失败：必填字段为空")
		return nil, status.Errorf(codes.InvalidArgument, "用户ID和词典ID不能为空")
	}

	// 创建收藏记录
	fav := &model.Favorite{
		ID:           uuid.New().String(),
		UserID:       req.UserId,
		DictionaryID: req.DictionaryId,
		MemoryDepth:  req.MemoryDepth,
		Model: model.Model{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := s.favoriteRepo.AddFavorite(fav)
	if err != nil {
		s.logger.Error("添加收藏失败",
			zap.String("userID", req.UserId),
			zap.Uint64("dictionaryID", req.DictionaryId),
			zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "添加收藏失败: %v", err)
	}

	// 转换响应
	response := &favorite.AddFavoriteResponse{
		Favorite: s.convertModelToProto(fav),
	}

	s.logger.Info("收藏添加成功", zap.String("favoriteID", fav.ID))
	return response, nil
}

// GetFavoritesByMemoryAsc 按memory升序查询收藏
func (s *FavoriteGRPCServer) GetFavoritesByMemoryAsc(ctx context.Context, req *favorite.GetFavoritesByMemoryAscRequest) (*favorite.GetFavoritesByMemoryAscResponse, error) {
	s.logger.Debug("按memory升序查询收藏",
		zap.String("userID", req.UserId),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	if req.UserId == "" {
		s.logger.Error("查询收藏失败：用户ID不能为空")
		return nil, status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}

	// 调用repository层
	favorites, err := s.favoriteRepo.GetFavoritesByMemoryAsc(req.UserId, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("按memory升序查询收藏失败",
			zap.String("userID", req.UserId),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "查询收藏失败: %v", err)
	}

	// 转换响应
	var protoFavorites []*favorite.Favorite
	for _, fav := range favorites {
		protoFavorites = append(protoFavorites, s.convertModelToProto(fav))
	}

	response := &favorite.GetFavoritesByMemoryAscResponse{
		Favorites: protoFavorites,
	}

	return response, nil
}

// GetFavoritesByStudyRecord 按学习记录查询收藏
func (s *FavoriteGRPCServer) GetFavoritesByStudyRecord(ctx context.Context, req *favorite.GetFavoritesByStudyRecordRequest) (*favorite.GetFavoritesByStudyRecordResponse, error) {
	s.logger.Debug("按收藏日志查询收藏",
		zap.String("userID", req.UserId),
		zap.String("result", req.Result),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	if req.UserId == "" {
		s.logger.Error("查询收藏失败：用户ID不能为空")
		return nil, status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}

	if req.Result == "" {
		s.logger.Error("查询收藏失败：学习结果参数不能为空")
		return nil, status.Errorf(codes.InvalidArgument, "学习结果参数不能为空")
	}

	// 验证result参数
	if req.Result != "remembered" && req.Result != "fuzzy" && req.Result != "strange" {
		s.logger.Error("查询收藏失败：学习结果参数无效", zap.String("result", req.Result))
		return nil, status.Errorf(codes.InvalidArgument, "学习结果参数无效: %s", req.Result)
	}

	// 调用repository层
	favorites, err := s.favoriteRepo.GetFavoritesByStudyRecord(req.UserId, req.Result, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("按收藏日志查询收藏失败",
			zap.String("userID", req.UserId),
			zap.String("result", req.Result),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "查询收藏失败: %v", err)
	}

	// 转换响应
	var protoFavorites []*favorite.Favorite
	for _, fav := range favorites {
		protoFavorites = append(protoFavorites, s.convertModelToProto(fav))
	}

	response := &favorite.GetFavoritesByStudyRecordResponse{
		Favorites: protoFavorites,
	}

	return response, nil
}

// GetFavoritesByMemoryDepth 按记忆深度查询收藏
func (s *FavoriteGRPCServer) GetFavoritesByMemoryDepth(ctx context.Context, req *favorite.GetFavoritesByMemoryDepthRequest) (*favorite.GetFavoritesByMemoryDepthResponse, error) {
	s.logger.Debug("按记忆深度查询收藏",
		zap.String("userID", req.UserId),
		zap.Uint64("memoryDepth", req.MemoryDepth),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	if req.UserId == "" {
		s.logger.Error("查询收藏失败：用户ID不能为空")
		return nil, status.Errorf(codes.InvalidArgument, "用户ID不能为空")
	}

	// 调用repository层
	favorites, err := s.favoriteRepo.GetFavoritesByMemoryDepth(req.UserId, req.MemoryDepth, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("按记忆深度查询收藏失败",
			zap.String("userID", req.UserId),
			zap.Uint64("memoryDepth", req.MemoryDepth),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "查询收藏失败: %v", err)
	}

	// 转换响应
	var protoFavorites []*favorite.Favorite
	for _, fav := range favorites {
		protoFavorites = append(protoFavorites, s.convertModelToProto(fav))
	}

	response := &favorite.GetFavoritesByMemoryDepthResponse{
		Favorites: protoFavorites,
	}

	return response, nil
}

// AddStudyRecord 添加学习记录
func (s *FavoriteGRPCServer) AddStudyRecord(ctx context.Context, req *favorite.AddStudyRecordRequest) (*favorite.AddStudyRecordResponse, error) {
	s.logger.Info("添加学习记录", zap.String("result", req.Result))

	// 验证result参数
	if req.Result != "remembered" && req.Result != "fuzzy" && req.Result != "strange" {
		s.logger.Error("添加学习记录失败：学习结果参数无效", zap.String("result", req.Result))
		return nil, status.Errorf(codes.InvalidArgument, "学习结果参数无效")
	}

	// 创建学习记录
	studyRecord := &model.StudyRecord{
		ID:     uuid.New().String(),
		Result: req.Result,
		Remark: req.Remark,
		Model: model.Model{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := s.favoriteRepo.AddStudyRecord(studyRecord)
	if err != nil {
		s.logger.Error("添加学习记录失败",
			zap.String("result", req.Result),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "添加学习记录失败: %v", err)
	}

	// 转换响应
	response := &favorite.AddStudyRecordResponse{
		StudyRecord: s.convertStudyRecordToProto(studyRecord),
	}

	s.logger.Info("学习记录添加成功", zap.String("studyRecordID", studyRecord.ID))
	return response, nil
}

// convertModelToProto 将模型转换为protobuf消息
func (s *FavoriteGRPCServer) convertModelToProto(fav *model.Favorite) *favorite.Favorite {
	protoFav := &favorite.Favorite{
		Id:           fav.ID,
		UserId:       fav.UserID,
		DictionaryId: fav.DictionaryID,
		MemoryDepth:  fav.MemoryDepth,
		CreatedAt:    timestamppb.New(fav.CreatedAt),
		UpdatedAt:    timestamppb.New(fav.UpdatedAt),
	}

	// 转换学习记录
	for _, record := range fav.FavoriteRecords {
		protoRecord := s.convertStudyRecordToProto(&record)
		protoFav.FavoriteRecords = append(protoFav.FavoriteRecords, protoRecord)
	}

	return protoFav
}

// convertStudyRecordToProto 将学习记录模型转换为protobuf消息
func (s *FavoriteGRPCServer) convertStudyRecordToProto(record *model.StudyRecord) *favorite.StudyRecord {
	return &favorite.StudyRecord{
		Id:        record.ID,
		Result:    record.Result,
		Remark:    record.Remark,
		CreatedAt: timestamppb.New(record.CreatedAt),
		UpdatedAt: timestamppb.New(record.UpdatedAt),
	}
}
