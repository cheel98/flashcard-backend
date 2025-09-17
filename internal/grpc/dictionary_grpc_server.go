package grpc

import (
	"context"
	"time"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/proto/generated/dictionary"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DictionaryGRPCServer gRPC词典服务实现
type DictionaryGRPCServer struct {
	dictionary.UnimplementedDictionaryServiceServer
	dictionaryRepo repository.DictionaryRepository
	logger         *zap.Logger
}

// NewDictionaryGRPCServer 创建新的gRPC词典服务
func NewDictionaryGRPCServer(dictionaryRepo repository.DictionaryRepository, logger *zap.Logger) *DictionaryGRPCServer {
	return &DictionaryGRPCServer{
		dictionaryRepo: dictionaryRepo,
		logger:         logger,
	}
}

// CreateDictionary 创建词典记录
func (s *DictionaryGRPCServer) CreateDictionary(ctx context.Context, req *dictionary.CreateDictionaryRequest) (*dictionary.CreateDictionaryResponse, error) {
	s.logger.Info("创建词典记录",
		zap.String("sourceLang", req.SourceLang),
		zap.String("targetLang", req.TargetLang),
		zap.String("sourceText", req.SourceText))

	// 验证必填字段
	if req.SourceLang == "" || req.TargetLang == "" || req.SourceText == "" || req.TranslatedText == "" {
		s.logger.Error("创建词典记录失败：必填字段为空")
		return nil, status.Errorf(codes.InvalidArgument, "源语言、目标语言、源文本和翻译文本不能为空")
	}

	// 创建词典记录
	dict := &model.Dictionary{
		SourceLang:      req.SourceLang,
		TargetLang:      req.TargetLang,
		SourceText:      req.SourceText,
		TranslatedText:  req.TranslatedText,
		PartOfSpeech:    req.PartOfSpeech,
		IPA:             req.Ipa,
		ExampleSentence: req.ExampleSentence,
		Model: model.Model{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := s.dictionaryRepo.CreateDictionary(dict)
	if err != nil {
		s.logger.Error("创建词典记录失败",
			zap.String("sourceLang", req.SourceLang),
			zap.String("targetLang", req.TargetLang),
			zap.String("sourceText", req.SourceText),
			zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "创建词典记录失败: %v", err)
	}

	// 转换响应
	response := &dictionary.CreateDictionaryResponse{
		Dictionary: s.convertModelToProto(dict),
	}

	s.logger.Info("词典记录创建成功", zap.Uint64("dictionaryID", dict.ID))
	return response, nil
}

// GetDictionaryByUniqueTranslation 根据唯一翻译信息查询词典
func (s *DictionaryGRPCServer) GetDictionaryByUniqueTranslation(ctx context.Context, req *dictionary.GetDictionaryByUniqueTranslationRequest) (*dictionary.GetDictionaryByUniqueTranslationResponse, error) {
	s.logger.Debug("根据唯一翻译信息查询词典",
		zap.String("sourceLang", req.SourceLang),
		zap.String("targetLang", req.TargetLang),
		zap.String("sourceText", req.SourceText))

	if req.SourceLang == "" || req.TargetLang == "" || req.SourceText == "" {
		s.logger.Error("查询词典失败：参数不能为空")
		return nil, status.Errorf(codes.InvalidArgument, "源语言、目标语言和源文本参数不能为空")
	}

	// 调用repository层
	dict, err := s.dictionaryRepo.GetDictionaryByUniqueTranslation(req.SourceLang, req.TargetLang, req.SourceText)
	if err != nil {
		s.logger.Error("查询词典记录失败",
			zap.String("sourceLang", req.SourceLang),
			zap.String("targetLang", req.TargetLang),
			zap.String("sourceText", req.SourceText),
			zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "查询词典记录失败: %v", err)
	}

	// 转换响应
	response := &dictionary.GetDictionaryByUniqueTranslationResponse{
		Dictionary: s.convertModelToProto(dict),
	}

	return response, nil
}

// convertModelToProto 将模型转换为protobuf消息
func (s *DictionaryGRPCServer) convertModelToProto(dict *model.Dictionary) *dictionary.Dictionary {
	protoDict := &dictionary.Dictionary{
		Id:              dict.ID,
		SourceLang:      dict.SourceLang,
		TargetLang:      dict.TargetLang,
		SourceText:      dict.SourceText,
		TranslatedText:  dict.TranslatedText,
		PartOfSpeech:    dict.PartOfSpeech,
		Ipa:             dict.IPA,
		ExampleSentence: dict.ExampleSentence,
		CreatedAt:       timestamppb.New(dict.CreatedAt),
		UpdatedAt:       timestamppb.New(dict.UpdatedAt),
	}

	// 转换音频数据
	for _, audio := range dict.Audios {
		protoAudio := &dictionary.DictionaryAudio{
			Id:           audio.ID,
			DictionaryId: audio.DictionaryID,
			AudioPath:    audio.AudioPath,
			Accent:       audio.Accent,
		}
		protoDict.Audios = append(protoDict.Audios, protoAudio)
	}

	// 转换元数据
	for _, metadata := range dict.Metadata {
		protoMetadata := &dictionary.DictionaryMetadata{
			Id:           metadata.ID,
			DictionaryId: metadata.DictionaryID,
			Key:          metadata.Key,
			Value:        metadata.Value,
		}
		protoDict.Metadata = append(protoDict.Metadata, protoMetadata)
	}

	return protoDict
}
