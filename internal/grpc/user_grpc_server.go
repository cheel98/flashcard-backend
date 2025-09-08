package grpc

import (
	"context"
	"strconv"

	"github.com/cheel98/flashcard-backend/internal/service"
	"github.com/cheel98/flashcard-backend/proto/generated/user"
)

// UserGRPCServer 实现 UserServiceServer 接口
type UserGRPCServer struct {
	user.UnimplementedUserServiceServer
	userService service.UserService
}

// NewUserGRPCServer 创建新的 UserGRPCServer 实例
func NewUserGRPCServer(userService service.UserService) *UserGRPCServer {
	return &UserGRPCServer{
		userService: userService,
	}
}

// Login 用户登录
func (s *UserGRPCServer) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	// 调用服务层进行登录
	tokenPair, err := s.userService.Login(req.Email, req.PasswordHash)
	if err != nil {
		return nil, err
	}

	return &user.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *UserGRPCServer) RefreshToken(ctx context.Context, req *user.RefreshTokenRequest) (*user.RefreshTokenResponse, error) {
	// 调用服务层刷新令牌
	accessToken, err := s.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &user.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

// Logout 用户登出
func (s *UserGRPCServer) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	// 调用服务层进行登出
	err := s.userService.Logout(req.UserId)
	if err != nil {
		return &user.LogoutResponse{
			Success: false,
		}, err
	}

	return &user.LogoutResponse{
		Success: true,
	}, nil
}

// GetUserByID 根据ID获取用户信息
func (s *UserGRPCServer) GetUserByID(ctx context.Context, req *user.GetUserByIDRequest) (*user.GetUserByIDResponse, error) {
	// 将字符串ID转换为整数
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, err
	}

	// 调用服务层获取用户信息
	userModel, err := s.userService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return &user.GetUserByIDResponse{
		User: convertUserModelToProto(userModel),
	}, nil
}

// GetUserSettings 获取用户设置
func (s *UserGRPCServer) GetUserSettings(ctx context.Context, req *user.GetUserSettingsRequest) (*user.GetUserSettingsResponse, error) {
	// 将字符串ID转换为整数
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, err
	}

	// 调用服务层获取用户设置
	settings, err := s.userService.GetUserSettings(userID)
	if err != nil {
		return nil, err
	}

	return &user.GetUserSettingsResponse{
		Settings: convertUserSettingsToProto(settings),
	}, nil
}

// GetUserPreferences 获取用户偏好
func (s *UserGRPCServer) GetUserPreferences(ctx context.Context, req *user.GetUserPreferencesRequest) (*user.GetUserPreferencesResponse, error) {
	// 将字符串ID转换为整数
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, err
	}

	// 调用服务层获取用户偏好
	preferences, err := s.userService.GetUserPreferences(userID)
	if err != nil {
		return nil, err
	}

	return &user.GetUserPreferencesResponse{
		Preferences: convertUserPreferencesToProto(preferences),
	}, nil
}

// GetUserLogs 获取用户日志
func (s *UserGRPCServer) GetUserLogs(ctx context.Context, req *user.GetUserLogsRequest) (*user.GetUserLogsResponse, error) {
	// 将字符串ID转换为整数
	userID, err := strconv.Atoi(req.UserId)
	if err != nil {
		return nil, err
	}

	// 调用服务层获取用户日志
	logs, err := s.userService.GetUserLogs(userID)
	if err != nil {
		return nil, err
	}

	// 转换日志列表
	protoLogs := make([]*user.UserLog, len(logs))
	for i, log := range logs {
		protoLogs[i] = convertUserLogToProto(log)
	}

	return &user.GetUserLogsResponse{
		Logs: protoLogs,
	}, nil
}

// 辅助函数：将用户模型转换为 protobuf 消息
func convertUserModelToProto(userModel interface{}) *user.User {
	// 这里需要根据实际的用户模型结构进行转换
	// 假设用户模型有 ID, Username, Email 等字段
	return &user.User{
		Id:       "1", // 需要根据实际模型字段转换
		Username: "username",
		Email:    "email@example.com",
	}
}

// 辅助函数：将用户设置转换为 protobuf 消息
func convertUserSettingsToProto(settings interface{}) *user.UserSettings {
	// 这里需要根据实际的用户设置结构进行转换
	return &user.UserSettings{
		Language: "zh-CN",
		Theme:    "light",
	}
}

// 辅助函数：将用户偏好转换为 protobuf 消息
func convertUserPreferencesToProto(preferences interface{}) *user.UserPreferences {
	// 这里需要根据实际的用户偏好结构进行转换
	return &user.UserPreferences{
		Difficulty: "medium",
		StudyMode:  "flashcard",
	}
}

// 辅助函数：将用户日志转换为 protobuf 消息
func convertUserLogToProto(log interface{}) *user.UserLog {
	// 这里需要根据实际的用户日志结构进行转换
	return &user.UserLog{
		Id:        "1",
		Action:    "login",
		Timestamp: "2024-01-01T00:00:00Z",
		Details:   "User logged in",
	}
}
