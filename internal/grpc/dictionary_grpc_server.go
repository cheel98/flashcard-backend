package grpc

import (
	"context"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/cheel98/flashcard-backend/proto/generated/dictionary"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DictionaryGRPCServer gRPC词典服务实现
type DictionaryGRPCServer struct {
	dictionary.UnimplementedDictionaryServiceServer
	dictionaryService service.DictionaryService
	logger            *zap.Logger
}

// NewDictionaryGRPCServer 创建新的gRPC词典服务
func NewDictionaryGRPCServer(dictionaryService service.DictionaryService, logger *zap.Logger) *DictionaryGRPCServer {
	return &DictionaryGRPCServer{
		dictionaryService: dictionaryService,
		logger:            logger,
	}
}

// CreateDictionary 创建词典记录
func (s *DictionaryGRPCServer) CreateDictionary(ctx context.Context, req *dictionary.CreateDictionaryRequest) (*dictionary.CreateDictionaryResponse, error) {
	s.logger.Info("gRPC CreateDictionary called",
		zap.String("source_lang", req.SourceLang),
		zap.String("target_lang", req.TargetLang),
		zap.String("source_text", req.SourceText))

	// 转换请求为服务层请求
	serviceReq := &service.CreateDictionaryRequest{
		SourceLang:      req.SourceLang,
		TargetLang:      req.TargetLang,
		SourceText:      req.SourceText,
		TranslatedText:  req.TranslatedText,
		PartOfSpeech:    req.PartOfSpeech,
		IPA:             req.Ipa,
		ExampleSentence: req.ExampleSentence,
	}

	// 调用服务层
	dict, err := s.dictionaryService.CreateDictionary(serviceReq)
	if err != nil {
		s.logger.Error("Failed to create dictionary", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "创建词典记录失败: %v", err)
	}

	// 转换响应
	response := &dictionary.CreateDictionaryResponse{
		Dictionary: s.convertModelToProto(dict),
	}

	s.logger.Info("Dictionary created successfully", zap.Uint64("id", dict.ID))
	return response, nil
}

// GetDictionaryByUniqueTranslation 根据唯一翻译信息查询词典
func (s *DictionaryGRPCServer) GetDictionaryByUniqueTranslation(ctx context.Context, req *dictionary.GetDictionaryByUniqueTranslationRequest) (*dictionary.GetDictionaryByUniqueTranslationResponse, error) {
	s.logger.Info("gRPC GetDictionaryByUniqueTranslation called",
		zap.String("source_lang", req.SourceLang),
		zap.String("target_lang", req.TargetLang),
		zap.String("source_text", req.SourceText))

	// 调用服务层
	dict, err := s.dictionaryService.GetDictionaryByUniqueTranslation(req.SourceLang, req.TargetLang, req.SourceText)
	if err != nil {
		s.logger.Error("Failed to get dictionary by unique translation", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "查询词典记录失败: %v", err)
	}

	// 转换响应
	response := &dictionary.GetDictionaryByUniqueTranslationResponse{
		Dictionary: s.convertModelToProto(dict),
	}

	return response, nil
}

// GetDictionaryWithDetails 获取词典详细信息
func (s *DictionaryGRPCServer) GetDictionaryWithDetails(ctx context.Context, req *dictionary.GetDictionaryWithDetailsRequest) (*dictionary.GetDictionaryWithDetailsResponse, error) {
	s.logger.Info("gRPC GetDictionaryWithDetails called", zap.Uint64("dictionary_id", req.DictionaryId))

	// 调用服务层
	dict, err := s.dictionaryService.GetDictionaryWithDetails(req.DictionaryId)
	if err != nil {
		s.logger.Error("Failed to get dictionary with details", zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "获取词典详细信息失败: %v", err)
	}

	// 转换响应
	response := &dictionary.GetDictionaryWithDetailsResponse{
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
