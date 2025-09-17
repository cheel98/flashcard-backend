package grpc

import (
	"context"
	"encoding/json"
	"github.com/cheel98/flashcard-backend/internal/config"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/internal/utils"
	"github.com/cheel98/flashcard-backend/internal/utils/authv3"
	"github.com/cheel98/flashcard-backend/proto/generated/translation"
	"log"
)

type YouDaoTranslationServer struct {
	translation.UnimplementedTranslationServer
	dicRepo   repository.DictionaryRepository
	url       string
	appKey    string
	appSecret string
}

func NewTranslationServerWithConfig(dictRepo repository.DictionaryRepository, config *config.Config) *YouDaoTranslationServer {
	return newTranslationServer(dictRepo, config.TransferConfig.URL, config.TransferConfig.AppKey, config.TransferConfig.AppSecret)
}

func newTranslationServer(dicRepo repository.DictionaryRepository, url, appKey, appSecret string) *YouDaoTranslationServer {
	return &YouDaoTranslationServer{
		dicRepo:   dicRepo,
		url:       url,
		appKey:    appKey,
		appSecret: appSecret,
	}
}

func (y *YouDaoTranslationServer) Translation(ctx context.Context, request *translation.TranslationRequest) (*translation.TranslationResponse, error) {
	params := make(map[string][]string)
	params["q"] = []string{request.Q}
	params["from"] = []string{request.From}
	params["to"] = []string{request.To}
	authv3.AddAuthParams(y.appKey, y.appSecret, params)
	res := &translation.TranslationResponse{}
	result := y.SendToEngine(params)
	log.Println(string(result))
	err := json.Unmarshal(result, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (y *YouDaoTranslationServer) SendToEngine(params map[string][]string) []byte {
	header := map[string][]string{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	res := utils.DoPost(y.url, header, params, "application/json")
	return res
}
