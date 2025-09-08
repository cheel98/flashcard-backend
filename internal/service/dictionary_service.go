package service

import (
	"fmt"
	"time"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"go.uber.org/zap"
)

// CreateDictionaryRequest 创建词典记录请求结构
type CreateDictionaryRequest struct {
	SourceLang      string `json:"source_lang"`
	TargetLang      string `json:"target_lang"`
	SourceText      string `json:"source_text"`
	TranslatedText  string `json:"translated_text"`
	PartOfSpeech    string `json:"part_of_speech,omitempty"`
	IPA             string `json:"ipa,omitempty"`
	ExampleSentence string `json:"example_sentence,omitempty"`
}

// DictionaryService 词典服务接口
type DictionaryService interface {
	CreateDictionary(req *CreateDictionaryRequest) (*model.Dictionary, error)
	GetDictionaryByUniqueTranslation(sourceLang, targetLang, sourceText string) (*model.Dictionary, error)
	GetDictionaryWithDetails(dictionaryID uint64) (*model.Dictionary, error)
}

// dictionaryService 词典服务实现
type dictionaryService struct {
	dictionaryRepo repository.DictionaryRepository
	logger         *zap.Logger
}

// NewDictionaryService 创建词典服务实例
func NewDictionaryService(dictionaryRepo repository.DictionaryRepository, logger *zap.Logger) DictionaryService {
	return &dictionaryService{
		dictionaryRepo: dictionaryRepo,
		logger:         logger,
	}
}

// CreateDictionary 创建词典记录
func (s *dictionaryService) CreateDictionary(req *CreateDictionaryRequest) (*model.Dictionary, error) {
	s.logger.Info("创建词典记录",
		zap.String("sourceLang", req.SourceLang),
		zap.String("targetLang", req.TargetLang),
		zap.String("sourceText", req.SourceText))

	// 验证必填字段
	if req.SourceLang == "" || req.TargetLang == "" || req.SourceText == "" || req.TranslatedText == "" {
		s.logger.Error("创建词典记录失败：必填字段为空")
		return nil, fmt.Errorf("源语言、目标语言、源文本和翻译文本不能为空")
	}

	// 创建词典记录
	dictionary := &model.Dictionary{
		SourceLang:      req.SourceLang,
		TargetLang:      req.TargetLang,
		SourceText:      req.SourceText,
		TranslatedText:  req.TranslatedText,
		PartOfSpeech:    req.PartOfSpeech,
		IPA:             req.IPA,
		ExampleSentence: req.ExampleSentence,
		Model: model.Model{
			CreatedAt: time.Now(),
			UpdateAt:  time.Now(),
		},
	}

	err := s.dictionaryRepo.CreateDictionary(dictionary)
	if err != nil {
		s.logger.Error("创建词典记录失败",
			zap.String("sourceLang", req.SourceLang),
			zap.String("targetLang", req.TargetLang),
			zap.String("sourceText", req.SourceText),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("词典记录创建成功", zap.Uint64("dictionaryID", dictionary.ID))
	return dictionary, nil
}

// GetDictionaryByUniqueTranslation 根据唯一翻译信息查询词典
func (s *dictionaryService) GetDictionaryByUniqueTranslation(sourceLang, targetLang, sourceText string) (*model.Dictionary, error) {
	s.logger.Debug("根据唯一翻译信息查询词典",
		zap.String("sourceLang", sourceLang),
		zap.String("targetLang", targetLang),
		zap.String("sourceText", sourceText))

	if sourceLang == "" || targetLang == "" || sourceText == "" {
		s.logger.Error("查询词典失败：参数不能为空")
		return nil, fmt.Errorf("源语言、目标语言和源文本参数不能为空")
	}

	dictionary, err := s.dictionaryRepo.GetDictionaryByUniqueTranslation(sourceLang, targetLang, sourceText)
	if err != nil {
		s.logger.Error("查询词典记录失败",
			zap.String("sourceLang", sourceLang),
			zap.String("targetLang", targetLang),
			zap.String("sourceText", sourceText),
			zap.Error(err))
		return nil, err
	}

	return dictionary, nil
}

// GetDictionaryWithDetails 获取词典详细信息
func (s *dictionaryService) GetDictionaryWithDetails(dictionaryID uint64) (*model.Dictionary, error) {
	s.logger.Debug("获取词典详细信息", zap.Uint64("dictionaryID", dictionaryID))

	dictionary, err := s.dictionaryRepo.GetDictionaryWithDetails(dictionaryID)
	if err != nil {
		s.logger.Error("获取词典详细信息失败",
			zap.Uint64("dictionaryID", dictionaryID),
			zap.Error(err))
		return nil, err
	}

	return dictionary, nil
}
