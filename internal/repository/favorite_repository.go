package repository

import (
	"errors"
	"github.com/cheel98/flashcard-backend/internal/model"
	"gorm.io/gorm"
)

// FavoriteRepository 收藏仓储接口
type FavoriteRepository interface {
	// AddFavorite 用户收藏单词
	AddFavorite(favorite *model.Favorite) error
	// GetFavoritesByMemoryAsc 按memory升序查询favorite
	GetFavoritesByMemoryAsc(userID string, limit, offset int) ([]*model.Favorite, error)
	// GetFavoritesByStudyRecord 按收藏日志查询Favorites
	GetFavoritesByStudyRecord(userID string, result string, limit, offset int) ([]*model.Favorite, error)
	// GetFavoritesByMemoryDepth 按记忆深度查询Favorites
	GetFavoritesByMemoryDepth(userID string, memoryDepth uint64, limit, offset int) ([]*model.Favorite, error)
	// AddStudyRecord 添加学习记录
	AddStudyRecord(record *model.StudyRecord) error
}

// favoriteRepository 收藏仓储实现
type favoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository 创建收藏仓储实例
func NewFavoriteRepository(db *gorm.DB) FavoriteRepository {
	return &favoriteRepository{
		db: db,
	}
}

// AddFavorite 用户收藏单词
func (r *favoriteRepository) AddFavorite(favorite *model.Favorite) error {
	// 检查是否已经收藏
	var existingFavorite model.Favorite
	err := r.db.Where("user_id = ? AND dictionary_id = ?", favorite.UserID, favorite.DictionaryID).First(&existingFavorite).Error
	if err == nil {
		return errors.New("该单词已经收藏")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 创建新的收藏记录
	err = r.db.Create(favorite).Error
	if err != nil {
		return err
	}
	return nil
}

// GetFavoritesByMemoryAsc 按memory升序查询favorite（关联查询dictionary表）
func (r *favoriteRepository) GetFavoritesByMemoryAsc(userID string, limit, offset int) ([]*model.Favorite, error) {
	var favorites []*model.Favorite
	err := r.db.Preload("DictionaryEnglishUS"). // 关联查询dictionary表
							Where("user_id = ?", userID).
							Order("memory_depth ASC").
							Limit(limit).
							Offset(offset).
							Find(&favorites).Error
	if err != nil {
		return nil, err
	}
	return favorites, nil
}

// GetFavoritesByStudyRecord 按收藏日志查询Favorites（关联查询dictionary表）
func (r *favoriteRepository) GetFavoritesByStudyRecord(userID string, result string, limit, offset int) ([]*model.Favorite, error) {
	var favorites []*model.Favorite

	// 先查询符合条件的StudyRecord，然后关联查询Favorite和Dictionary
	subQuery := r.db.Model(&model.StudyRecord{}).Select("id").Where("result = ?", result)

	err := r.db.Preload("DictionaryEnglishUS"). // 关联查询dictionary表
							Preload("FavoriteRecords", "result = ?", result). // 预加载符合条件的学习记录
							Where("user_id = ? AND id IN (?)", userID, subQuery).
							Limit(limit).
							Offset(offset).
							Find(&favorites).Error
	if err != nil {
		return nil, err
	}
	return favorites, nil
}

// GetFavoritesByMemoryDepth 按记忆深度查询Favorites（关联查询dictionary表）
func (r *favoriteRepository) GetFavoritesByMemoryDepth(userID string, memoryDepth uint64, limit, offset int) ([]*model.Favorite, error) {
	var favorites []*model.Favorite
	err := r.db.Preload("DictionaryEnglishUS"). // 关联查询dictionary表
							Where("user_id = ? AND memory_depth = ?", userID, memoryDepth).
							Order("created_at DESC").
							Limit(limit).
							Offset(offset).
							Find(&favorites).Error
	if err != nil {
		return nil, err
	}
	return favorites, nil
}

// AddStudyRecord 添加学习记录
func (r *favoriteRepository) AddStudyRecord(record *model.StudyRecord) error {
	err := r.db.Create(record).Error
	if err != nil {
		return err
	}
	return nil
}
