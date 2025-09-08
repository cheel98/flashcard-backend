package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// FavoriteHandler 收藏处理器
type FavoriteHandler struct {
	favoriteService service.FavoriteService
	logger          *zap.Logger
}

// NewFavoriteHandler 创建收藏处理器
func NewFavoriteHandler(favoriteService service.FavoriteService, logger *zap.Logger) *FavoriteHandler {
	return &FavoriteHandler{
		favoriteService: favoriteService,
		logger:          logger,
	}
}

// AddStudyRecordRequest 添加学习记录请求结构
type AddStudyRecordRequest struct {
	Result string `json:"result"`
	Remark string `json:"remark,omitempty"`
}

// AddFavorite 添加收藏接口
func (h *FavoriteHandler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	var req service.AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("解析收藏添加请求失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	favorite, err := h.favoriteService.AddFavorite(&req)
	if err != nil {
		h.logger.Error("添加收藏失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeSuccessResponse(w, favorite)
}

// GetFavoritesByMemoryAsc 按memory升序查询favorite接口
func (h *FavoriteHandler) GetFavoritesByMemoryAsc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	// 获取分页参数
	limit, offset := h.getPaginationParams(r)

	favorites, err := h.favoriteService.GetFavoritesByMemoryAsc(userID, limit, offset)
	if err != nil {
		h.logger.Error("按memory升序查询收藏失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "查询收藏失败")
		return
	}

	h.writeSuccessResponse(w, favorites)
}

// GetFavoritesByStudyRecord 按收藏日志查询Favorites接口
func (h *FavoriteHandler) GetFavoritesByStudyRecord(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	result := r.URL.Query().Get("result")

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	if result == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "学习结果参数不能为空")
		return
	}

	// 验证result参数
	if result != "remembered" && result != "fuzzy" && result != "strange" {
		h.writeErrorResponse(w, http.StatusBadRequest, "学习结果参数无效")
		return
	}

	// 获取分页参数
	limit, offset := h.getPaginationParams(r)

	favorites, err := h.favoriteService.GetFavoritesByStudyRecord(userID, result, limit, offset)
	if err != nil {
		h.logger.Error("按收藏日志查询收藏失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "查询收藏失败")
		return
	}

	h.writeSuccessResponse(w, favorites)
}

// GetFavoritesByMemoryDepth 按记忆深度查询Favorites接口
func (h *FavoriteHandler) GetFavoritesByMemoryDepth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]
	memoryDepthStr := r.URL.Query().Get("memory_depth")

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	if memoryDepthStr == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "记忆深度参数不能为空")
		return
	}

	memoryDepth, err := strconv.ParseUint(memoryDepthStr, 10, 64)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "记忆深度参数格式错误")
		return
	}

	// 获取分页参数
	limit, offset := h.getPaginationParams(r)

	favorites, err := h.favoriteService.GetFavoritesByMemoryDepth(userID, memoryDepth, limit, offset)
	if err != nil {
		h.logger.Error("按记忆深度查询收藏失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "查询收藏失败")
		return
	}

	h.writeSuccessResponse(w, favorites)
}

// AddStudyRecord 添加学习记录接口
func (h *FavoriteHandler) AddStudyRecord(w http.ResponseWriter, r *http.Request) {
	var req service.AddStudyRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("解析学习记录请求失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	studyRecord, err := h.favoriteService.AddStudyRecord(&req)
	if err != nil {
		h.logger.Error("添加学习记录失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "添加学习记录失败")
		return
	}

	h.writeSuccessResponse(w, studyRecord)
}

// RegisterRoutes 注册收藏相关路由
func (h *FavoriteHandler) RegisterRoutes(router *mux.Router) {
	favoriteRouter := router.PathPrefix("/api/v1/favorites").Subrouter()

	favoriteRouter.HandleFunc("", h.AddFavorite).Methods("POST")
	favoriteRouter.HandleFunc("/users/{userID}/memory-asc", h.GetFavoritesByMemoryAsc).Methods("GET")
	favoriteRouter.HandleFunc("/users/{userID}/study-record", h.GetFavoritesByStudyRecord).Methods("GET")
	favoriteRouter.HandleFunc("/users/{userID}/memory-depth", h.GetFavoritesByMemoryDepth).Methods("GET")
	favoriteRouter.HandleFunc("/study-records", h.AddStudyRecord).Methods("POST")
}

// getPaginationParams 获取分页参数
func (h *FavoriteHandler) getPaginationParams(r *http.Request) (limit, offset int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit = 10 // 默认值
	offset = 0 // 默认值

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

// writeSuccessResponse 写入成功响应
func (h *FavoriteHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
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
func (h *FavoriteHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
