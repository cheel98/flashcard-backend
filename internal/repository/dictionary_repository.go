package repository

import (
	"errors"
	"github.com/cheel98/flashcard-backend/internal/model"
	"gorm.io/gorm"
)

// DictionaryRepository 词典仓储接口
type DictionaryRepository interface {
	// CreateDictionary 新增词典记录
	CreateDictionary(dictionary *model.Dictionary) error
	// GetDictionaryByUniqueTranslation 根据idx_unique_translation信息查询dictionary
	GetDictionaryByUniqueTranslation(sourceLang, targetLang, sourceText string) (*model.Dictionary, error)
	// CreateDictionaryAudio 创建词典音频记录
	CreateDictionaryAudio(audio *model.DictionaryAudio) error
	// CreateDictionaryMetadata 创建词典元数据记录
	CreateDictionaryMetadata(metadata *model.DictionaryMetadata) error
	// GetDictionaryWithDetails 获取词典详细信息（包含音频和元数据）
	GetDictionaryWithDetails(dictionaryID uint64) (*model.Dictionary, error)
}

// dictionaryRepository 词典仓储实现
type dictionaryRepository struct {
	db *gorm.DB
}

// NewDictionaryRepository 创建词典仓储实例
func NewDictionaryRepository(db *gorm.DB) DictionaryRepository {
	return &dictionaryRepository{
		db: db,
	}
}

// CreateDictionary 新增词典记录
func (r *dictionaryRepository) CreateDictionary(dictionary *model.Dictionary) error {
	// 检查是否已存在相同的翻译记录（根据唯一索引）
	var existingDict model.Dictionary
	err := r.db.Where("source_lang = ? AND target_lang = ? AND source_text = ?",
		dictionary.SourceLang, dictionary.TargetLang, dictionary.SourceText).First(&existingDict).Error

	if err == nil {
		return errors.New("该翻译记录已存在")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	// 创建新的词典记录
	err = r.db.Create(dictionary).Error
	if err != nil {
		return err
	}
	return nil
}

// GetDictionaryByUniqueTranslation 根据idx_unique_translation信息查询dictionary
func (r *dictionaryRepository) GetDictionaryByUniqueTranslation(sourceLang, targetLang, sourceText string) (*model.Dictionary, error) {
	var dictionary model.Dictionary
	err := r.db.Where("source_lang = ? AND target_lang = ? AND source_text = ?",
		sourceLang, targetLang, sourceText).First(&dictionary).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("词典记录不存在")
		}
		return nil, err
	}
	return &dictionary, nil
}

// CreateDictionaryAudio 创建词典音频记录
func (r *dictionaryRepository) CreateDictionaryAudio(audio *model.DictionaryAudio) error {
	err := r.db.Create(audio).Error
	if err != nil {
		return err
	}
	return nil
}

// CreateDictionaryMetadata 创建词典元数据记录
func (r *dictionaryRepository) CreateDictionaryMetadata(metadata *model.DictionaryMetadata) error {
	err := r.db.Create(metadata).Error
	if err != nil {
		return err
	}
	return nil
}

// GetDictionaryWithDetails 获取词典详细信息（包含音频和元数据）
func (r *dictionaryRepository) GetDictionaryWithDetails(dictionaryID uint64) (*model.Dictionary, error) {
	var dictionary model.Dictionary
	err := r.db.Preload("Audios").Preload("Metadata").Where("id = ?", dictionaryID).First(&dictionary).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("词典记录不存在")
		}
		return nil, err
	}
	return &dictionary, nil
}
