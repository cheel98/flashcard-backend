package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AddFavoriteRequest 添加收藏请求结构
type AddFavoriteRequest struct {
	UserID       string `json:"user_id"`
	DictionaryID uint64 `json:"dictionary_id"`
	MemoryDepth  uint64 `json:"memory_depth"`
}

// AddStudyRecordRequest 添加学习记录请求结构
type AddStudyRecordRequest struct {
	Result string `json:"result"`
	Remark string `json:"remark,omitempty"`
}

// FavoriteService 收藏服务接口
type FavoriteService interface {
	AddFavorite(req *AddFavoriteRequest) (*model.Favorite, error)
	GetFavoritesByMemoryAsc(userID string, limit, offset int) ([]*model.Favorite, error)
	GetFavoritesByStudyRecord(userID, result string, limit, offset int) ([]*model.Favorite, error)
	GetFavoritesByMemoryDepth(userID string, memoryDepth uint64, limit, offset int) ([]*model.Favorite, error)
	AddStudyRecord(req *AddStudyRecordRequest) (*model.StudyRecord, error)
	GetPaginationParams(limitStr, offsetStr string) (int, int)
}

// favoriteService 收藏服务实现
type favoriteService struct {
	favoriteRepo repository.FavoriteRepository
	logger       *zap.Logger
}

// NewFavoriteService 创建收藏服务实例
func NewFavoriteService(favoriteRepo repository.FavoriteRepository, logger *zap.Logger) FavoriteService {
	return &favoriteService{
		favoriteRepo: favoriteRepo,
		logger:       logger,
	}
}

// AddFavorite 添加收藏
func (s *favoriteService) AddFavorite(req *AddFavoriteRequest) (*model.Favorite, error) {
	s.logger.Info("添加收藏",
		zap.String("userID", req.UserID),
		zap.Uint64("dictionaryID", req.DictionaryID))

	// 验证必填字段
	if req.UserID == "" || req.DictionaryID == 0 {
		s.logger.Error("添加收藏失败：必填字段为空")
		return nil, fmt.Errorf("用户ID和词典ID不能为空")
	}

	// 创建收藏记录
	favorite := &model.Favorite{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		DictionaryID: req.DictionaryID,
		MemoryDepth:  req.MemoryDepth,
		Model: model.Model{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := s.favoriteRepo.AddFavorite(favorite)
	if err != nil {
		s.logger.Error("添加收藏失败",
			zap.String("userID", req.UserID),
			zap.Uint64("dictionaryID", req.DictionaryID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("收藏添加成功", zap.String("favoriteID", favorite.ID))
	return favorite, nil
}

// GetFavoritesByMemoryAsc 按memory升序查询favorite
func (s *favoriteService) GetFavoritesByMemoryAsc(userID string, limit, offset int) ([]*model.Favorite, error) {
	s.logger.Debug("按memory升序查询收藏",
		zap.String("userID", userID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	if userID == "" {
		s.logger.Error("查询收藏失败：用户ID不能为空")
		return nil, fmt.Errorf("用户ID不能为空")
	}

	favorites, err := s.favoriteRepo.GetFavoritesByMemoryAsc(userID, limit, offset)
	if err != nil {
		s.logger.Error("按memory升序查询收藏失败",
			zap.String("userID", userID),
			zap.Error(err))
		return nil, err
	}

	return favorites, nil
}

// GetFavoritesByStudyRecord 按收藏日志查询Favorites
func (s *favoriteService) GetFavoritesByStudyRecord(userID, result string, limit, offset int) ([]*model.Favorite, error) {
	s.logger.Debug("按收藏日志查询收藏",
		zap.String("userID", userID),
		zap.String("result", result),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	if userID == "" {
		s.logger.Error("查询收藏失败：用户ID不能为空")
		return nil, fmt.Errorf("用户ID不能为空")
	}

	if result == "" {
		s.logger.Error("查询收藏失败：学习结果参数不能为空")
		return nil, fmt.Errorf("学习结果参数不能为空")
	}

	// 验证result参数
	if result != "remembered" && result != "fuzzy" && result != "strange" {
		s.logger.Error("查询收藏失败：学习结果参数无效", zap.String("result", result))
		return nil, fmt.Errorf("学习结果参数无效")
	}

	favorites, err := s.favoriteRepo.GetFavoritesByStudyRecord(userID, result, limit, offset)
	if err != nil {
		s.logger.Error("按收藏日志查询收藏失败",
			zap.String("userID", userID),
			zap.String("result", result),
			zap.Error(err))
		return nil, err
	}

	return favorites, nil
}

// GetFavoritesByMemoryDepth 按记忆深度查询Favorites
func (s *favoriteService) GetFavoritesByMemoryDepth(userID string, memoryDepth uint64, limit, offset int) ([]*model.Favorite, error) {
	s.logger.Debug("按记忆深度查询收藏",
		zap.String("userID", userID),
		zap.Uint64("memoryDepth", memoryDepth),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	if userID == "" {
		s.logger.Error("查询收藏失败：用户ID不能为空")
		return nil, fmt.Errorf("用户ID不能为空")
	}

	favorites, err := s.favoriteRepo.GetFavoritesByMemoryDepth(userID, memoryDepth, limit, offset)
	if err != nil {
		s.logger.Error("按记忆深度查询收藏失败",
			zap.String("userID", userID),
			zap.Uint64("memoryDepth", memoryDepth),
			zap.Error(err))
		return nil, err
	}

	return favorites, nil
}

// AddStudyRecord 添加学习记录
func (s *favoriteService) AddStudyRecord(req *AddStudyRecordRequest) (*model.StudyRecord, error) {
	s.logger.Info("添加学习记录", zap.String("result", req.Result))

	// 验证result参数
	if req.Result != "remembered" && req.Result != "fuzzy" && req.Result != "strange" {
		s.logger.Error("添加学习记录失败：学习结果参数无效", zap.String("result", req.Result))
		return nil, fmt.Errorf("学习结果参数无效")
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
		return nil, err
	}

	s.logger.Info("学习记录添加成功", zap.String("studyRecordID", studyRecord.ID))
	return studyRecord, nil
}

// GetPaginationParams 获取分页参数
func (s *favoriteService) GetPaginationParams(limitStr, offsetStr string) (int, int) {
	limit := 10 // 默认值
	offset := 0 // 默认值

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	return limit, offset
}
