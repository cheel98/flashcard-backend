package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// DictionaryHandler 词典处理器
type DictionaryHandler struct {
	dictionaryService service.DictionaryService
	logger            *zap.Logger
}

// NewDictionaryHandler 创建词典处理器
func NewDictionaryHandler(dictionaryService service.DictionaryService, logger *zap.Logger) *DictionaryHandler {
	return &DictionaryHandler{
		dictionaryService: dictionaryService,
		logger:            logger,
	}
}

// CreateDictionary 新增词典记录接口
func (h *DictionaryHandler) CreateDictionary(w http.ResponseWriter, r *http.Request) {
	var req service.CreateDictionaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("解析词典创建请求失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	dictionary, err := h.dictionaryService.CreateDictionary(&req)
	if err != nil {
		h.logger.Error("创建词典记录失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, dictionary)
}

// GetDictionaryByUniqueTranslation 根据idx_unique_translation信息查询dictionary接口
func (h *DictionaryHandler) GetDictionaryByUniqueTranslation(w http.ResponseWriter, r *http.Request) {
	sourceLang := r.URL.Query().Get("source_lang")
	targetLang := r.URL.Query().Get("target_lang")
	sourceText := r.URL.Query().Get("source_text")

	dictionary, err := h.dictionaryService.GetDictionaryByUniqueTranslation(sourceLang, targetLang, sourceText)
	if err != nil {
		h.logger.Error("查询词典记录失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, dictionary)
}

// GetDictionaryWithDetails 获取词典详细信息接口
func (h *DictionaryHandler) GetDictionaryWithDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dictionaryIDStr := vars["dictionaryID"]

	if dictionaryIDStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "词典ID不能为空")
		return
	}

	dictionaryID, err := strconv.ParseUint(dictionaryIDStr, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "词典ID格式错误")
		return
	}

	dictionary, err := h.dictionaryService.GetDictionaryWithDetails(dictionaryID)
	if err != nil {
		h.logger.Error("获取词典详细信息失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, dictionary)
}

// RegisterRoutes 注册词典相关路由
func (h *DictionaryHandler) RegisterRoutes(router *mux.Router) {
	dictionaryRouter := router.PathPrefix("/api/v1/dictionaries").Subrouter()

	dictionaryRouter.HandleFunc("", h.CreateDictionary).Methods("POST")
	dictionaryRouter.HandleFunc("/search", h.GetDictionaryByUniqueTranslation).Methods("GET")
	dictionaryRouter.HandleFunc("/{dictionaryID}", h.GetDictionaryWithDetails).Methods("GET")
}

// writeSuccessResponse 写入成功响应
func (h *DictionaryHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := Response{
		Code:    200,
		Message: "success",
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse 写入错误响应
func (h *DictionaryHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
