package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService service.UserService
	logger      *zap.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Login 用户登录接口
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("解析登录请求失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "请求参数错误")
		return
	}

	user, err := h.userService.Login(req.Email, req.PasswordHash)
	if err != nil {
		h.logger.Error("用户登录失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.writeSuccessResponse(w, user)
}

// GetUserInfo 获取用户基本信息接口
func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		h.logger.Error("获取用户信息失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, user)
}

// GetUserSettings 获取用户设置接口
func (h *UserHandler) GetUserSettings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	settings, err := h.userService.GetUserSettings(userID)
	if err != nil {
		h.logger.Error("获取用户设置失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, settings)
}

// GetUserPreferences 获取用户个人喜好接口
func (h *UserHandler) GetUserPreferences(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	preferences, err := h.userService.GetUserPreferences(userID)
	if err != nil {
		h.logger.Error("获取用户喜好失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	h.writeSuccessResponse(w, preferences)
}

// GetUserLogs 获取用户操作日志接口
func (h *UserHandler) GetUserLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "用户ID不能为空")
		return
	}

	// 获取分页参数
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

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

	logs, err := h.userService.GetUserLogs(userID, limit, offset)
	if err != nil {
		h.logger.Error("获取用户日志失败", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "获取用户日志失败")
		return
	}

	h.writeSuccessResponse(w, logs)
}

// RegisterRoutes 注册用户相关路由
func (h *UserHandler) RegisterRoutes(router *mux.Router) {
	userRouter := router.PathPrefix("/api/v1/users").Subrouter()

	userRouter.HandleFunc("/login", h.Login).Methods("POST")
	userRouter.HandleFunc("/{userID}", h.GetUserInfo).Methods("GET")
	userRouter.HandleFunc("/{userID}/settings", h.GetUserSettings).Methods("GET")
	userRouter.HandleFunc("/{userID}/preferences", h.GetUserPreferences).Methods("GET")
	userRouter.HandleFunc("/{userID}/logs", h.GetUserLogs).Methods("GET")
}

// writeSuccessResponse 写入成功响应
func (h *UserHandler) writeSuccessResponse(w http.ResponseWriter, data interface{}) {
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
func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Code:    statusCode,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
