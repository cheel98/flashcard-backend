package grpc

import (
	"context"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/cheel98/flashcard-backend/proto/generated/favorite"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FavoriteGRPCServer gRPC收藏服务实现
type FavoriteGRPCServer struct {
	favorite.UnimplementedFavoriteServiceServer
	favoriteService service.FavoriteService
	logger          *zap.Logger
}

// NewFavoriteGRPCServer 创建新的gRPC收藏服务
func NewFavoriteGRPCServer(favoriteService service.FavoriteService, logger *zap.Logger) *FavoriteGRPCServer {
	return &FavoriteGRPCServer{
		favoriteService: favoriteService,
		logger:          logger,
	}
}

// AddFavorite 添加收藏
func (s *FavoriteGRPCServer) AddFavorite(ctx context.Context, req *favorite.AddFavoriteRequest) (*favorite.AddFavoriteResponse, error) {
	s.logger.Info("gRPC AddFavorite called",
		zap.String("user_id", req.UserId),
		zap.Uint64("dictionary_id", req.DictionaryId))

	// 转换请求为服务层请求
	serviceReq := &service.AddFavoriteRequest{
		UserID:       req.UserId,
		DictionaryID: req.DictionaryId,
		MemoryDepth:  req.MemoryDepth,
	}

	// 调用服务层
	fav, err := s.favoriteService.AddFavorite(serviceReq)
	if err != nil {
		s.logger.Error("Failed to add favorite", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "添加收藏失败: %v", err)
	}

	// 转换响应
	response := &favorite.AddFavoriteResponse{
		Favorite: s.convertModelToProto(fav),
	}

	s.logger.Info("Favorite added successfully", zap.String("favorite_id", fav.ID))
	return response, nil
}

// GetFavoritesByMemoryAsc 按memory升序查询收藏
func (s *FavoriteGRPCServer) GetFavoritesByMemoryAsc(ctx context.Context, req *favorite.GetFavoritesByMemoryAscRequest) (*favorite.GetFavoritesByMemoryAscResponse, error) {
	s.logger.Info("gRPC GetFavoritesByMemoryAsc called",
		zap.String("user_id", req.UserId),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	// 调用服务层
	favorites, err := s.favoriteService.GetFavoritesByMemoryAsc(req.UserId, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("Failed to get favorites by memory asc", zap.Error(err))
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
	s.logger.Info("gRPC GetFavoritesByStudyRecord called",
		zap.String("user_id", req.UserId),
		zap.String("result", req.Result),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	// 验证result参数
	if req.Result != "remembered" && req.Result != "fuzzy" && req.Result != "strange" {
		return nil, status.Errorf(codes.InvalidArgument, "学习结果参数无效: %s", req.Result)
	}

	// 调用服务层
	favorites, err := s.favoriteService.GetFavoritesByStudyRecord(req.UserId, req.Result, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("Failed to get favorites by study record", zap.Error(err))
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
	s.logger.Info("gRPC GetFavoritesByMemoryDepth called",
		zap.String("user_id", req.UserId),
		zap.Uint64("memory_depth", req.MemoryDepth),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	// 调用服务层
	favorites, err := s.favoriteService.GetFavoritesByMemoryDepth(req.UserId, req.MemoryDepth, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("Failed to get favorites by memory depth", zap.Error(err))
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
	s.logger.Info("gRPC AddStudyRecord called",
		zap.String("result", req.Result),
		zap.String("remark", req.Remark))

	// 转换请求为服务层请求
	serviceReq := &service.AddStudyRecordRequest{
		Result: req.Result,
		Remark: req.Remark,
	}

	// 调用服务层
	studyRecord, err := s.favoriteService.AddStudyRecord(serviceReq)
	if err != nil {
		s.logger.Error("Failed to add study record", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "添加学习记录失败: %v", err)
	}

	// 转换响应
	response := &favorite.AddStudyRecordResponse{
		StudyRecord: s.convertStudyRecordToProto(studyRecord),
	}

	s.logger.Info("Study record added successfully", zap.String("record_id", studyRecord.ID))
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
