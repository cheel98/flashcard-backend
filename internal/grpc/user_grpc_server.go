package grpc

import (
	"context"

	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/pkg/email"
	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"github.com/cheel98/flashcard-backend/pkg/redis"
	"github.com/cheel98/flashcard-backend/proto/generated/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// UserGRPCServer gRPC用户服务实现
type UserGRPCServer struct {
	user.UnimplementedUserServiceServer
	userRepo     repository.UserRepository
	jwtManager   *jwt.JWTManager
	redisClient  *redis.RedisClient
	emailService *email.EmailService
	logger       *zap.Logger
}

// NewUserGRPCServer 创建新的gRPC用户服务
func NewUserGRPCServer(userRepo repository.UserRepository, jwtManager *jwt.JWTManager, redisClient *redis.RedisClient, emailService *email.EmailService, logger *zap.Logger) *UserGRPCServer {
	return &UserGRPCServer{
		userRepo:     userRepo,
		jwtManager:   jwtManager,
		redisClient:  redisClient,
		emailService: emailService,
		logger:       logger,
	}
}
func (s *UserGRPCServer) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	s.logger.Info("Register", zap.String("email", req.Email), zap.String("name", req.Name))

	// 创建用户
	user_, err := s.userRepo.Create(&model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.PasswordHash,
	})
	if err != nil {
		s.logger.Error("用户注册失败", zap.String("email", req.Email), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "用户注册失败: %v", err)
	}

	// 注册成功后删除Redis中的验证码
	err = s.redisClient.DeleteCaptcha(ctx, req.Email)
	if err != nil {
		s.logger.Warn("删除验证码失败", zap.String("email", req.Email), zap.Error(err))
	}

	s.logger.Info("用户注册成功", zap.String("email", req.Email), zap.String("userID", user_.ID))
	return &user.RegisterResponse{
		UserId: user_.ID,
	}, nil
}

var FailedBool = &user.BoolResponse{Success: false}
var SuccessBool = &user.BoolResponse{Success: true}

func (s *UserGRPCServer) VerifyCaptcha(ctx context.Context, request *user.CaptchaRequest) (*user.BoolResponse, error) {
	captcha, err := s.redisClient.GetCaptcha(ctx, request.Email)
	if err != nil {
		return FailedBool, nil
	}
	if captcha == request.GetCode() {
		return SuccessBool, nil
	}
	return FailedBool, nil
}

func (s *UserGRPCServer) SendEmailCaptcha(ctx context.Context, request *user.SendCaptchaRequest) (*user.BoolResponse, error) {
	// 1.生成验证码
	captcha, err := s.emailService.GenerateCaptcha()
	if err != nil {
		return FailedBool, err
	}
	// 2.将验证码存在redis中，过期时间为5分钟，key为邮箱，value为captcha
	err = s.redisClient.SetCaptcha(ctx, request.Email, captcha)
	if err != nil {
		return FailedBool, err
	}
	// 3.将验证码发送到用户邮箱
	err = s.emailService.SendCaptcha(request.Email, captcha)
	if err != nil {
		return FailedBool, err
	}
	return SuccessBool, nil
}

// Login 用户登录
func (s *UserGRPCServer) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	s.logger.Info("gRPC Login called",
		zap.String("email", req.Email))

	// 直接调用repository层进行用户验证
	user_, err := s.userRepo.Login(req.Email, req.PasswordHash)
	if err != nil {
		s.logger.Error("用户登录失败", zap.String("email", req.Email), zap.Error(err))
		return nil, status.Errorf(codes.Unauthenticated, "登录失败: %v", err)
	}

	// 生成token对
	tokenPair, err := s.jwtManager.GenerateTokenPair(user_.ID, user_.Email)
	if err != nil {
		s.logger.Error("生成token失败", zap.String("userID", user_.ID), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "生成token失败: %v", err)
	}

	// 保存refresh token到数据库
	err = s.userRepo.SaveRefreshToken(user_.ID, tokenPair.RefreshToken)
	if err != nil {
		s.logger.Error("保存refresh token失败", zap.String("userID", user_.ID), zap.Error(err))
		return nil, status.Errorf(codes.Internal, "保存refresh token失败: %v", err)
	}

	s.logger.Info("用户登录成功", zap.String("email", req.Email), zap.String("userID", user_.ID))
	return &user.LoginResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshToken 刷新访问令牌
func (s *UserGRPCServer) RefreshToken(ctx context.Context, req *user.RefreshTokenRequest) (*user.RefreshTokenResponse, error) {
	s.logger.Debug("刷新访问令牌")
	// 验证refresh token
	user_, err := s.userRepo.GetUserByRefreshToken(req.RefreshToken)
	if err != nil {
		s.logger.Error("无效的refresh token", zap.Error(err))
		return nil, err
	}

	// 生成新的access token
	accessToken, err := s.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		s.logger.Error("刷新access token失败", zap.String("userID", user_.ID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("access token刷新成功", zap.String("userID", user_.ID))
	return &user.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}

// Logout 用户登出
func (s *UserGRPCServer) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	s.logger.Info("gRPC Logout called",
		zap.String("user_id", req.UserId))
	// 清除refresh token
	err := s.userRepo.ClearRefreshToken(req.UserId)
	if err != nil {
		s.logger.Error("清除refresh token失败", zap.String("userID", req.UserId), zap.Error(err))
		return nil, err
	}

	s.logger.Info("用户登出成功", zap.String("userID", req.UserId))
	return &user.LogoutResponse{
		Success: true,
	}, nil
}

// GetUserByID 获取用户信息
func (s *UserGRPCServer) GetUserByID(ctx context.Context, req *user.GetUserByIDRequest) (*user.GetUserByIDResponse, error) {
	s.logger.Debug("获取用户信息", zap.String("userID", req.UserId))

	userModel, err := s.userRepo.GetUserByID(req.UserId)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "获取用户信息失败: %v", err)
	}

	return &user.GetUserByIDResponse{
		User: s.convertUserModelToProto(userModel),
	}, nil
}

// GetUserSettings 获取用户设置
func (s *UserGRPCServer) GetUserSettings(ctx context.Context, req *user.GetUserSettingsRequest) (*user.GetUserSettingsResponse, error) {
	s.logger.Debug("获取用户设置", zap.String("userID", req.UserId))

	userSettings, err := s.userRepo.GetUserSettings(req.UserId)
	if err != nil {
		s.logger.Error("获取用户设置失败", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "获取用户设置失败: %v", err)
	}

	return &user.GetUserSettingsResponse{
		UserSettings: s.convertUserSettingsModelToProto(userSettings),
	}, nil
}

// GetUserPreferences 获取用户偏好
func (s *UserGRPCServer) GetUserPreferences(ctx context.Context, req *user.GetUserPreferencesRequest) (*user.GetUserPreferencesResponse, error) {
	s.logger.Debug("获取用户偏好设置", zap.String("userID", req.UserId))

	userPreferences, err := s.userRepo.GetUserPreferences(req.UserId)
	if err != nil {
		s.logger.Error("获取用户偏好设置失败", zap.String("userID", req.UserId), zap.Error(err))
		return nil, status.Errorf(codes.NotFound, "获取用户偏好失败: %v", err)
	}

	return &user.GetUserPreferencesResponse{
		UserPreferences: s.convertUserPreferencesModelToProto(userPreferences),
	}, nil
}

// GetUserLogs 获取用户日志
func (s *UserGRPCServer) GetUserLogs(ctx context.Context, req *user.GetUserLogsRequest) (*user.GetUserLogsResponse, error) {
	s.logger.Debug("获取用户操作日志",
		zap.String("userID", req.UserId),
		zap.Int32("limit", req.Limit),
		zap.Int32("offset", req.Offset))

	userLogs, err := s.userRepo.GetUserLogs(req.UserId, int(req.Limit), int(req.Offset))
	if err != nil {
		s.logger.Error("获取用户操作日志失败",
			zap.String("userID", req.UserId),
			zap.Error(err))
		return nil, status.Errorf(codes.Internal, "获取用户日志失败: %v", err)
	}

	// 转换为proto格式
	protoLogs := make([]*user.UserLogs, len(userLogs))
	for i, log := range userLogs {
		protoLogs[i] = s.convertUserLogsModelToProto(log)
	}

	return &user.GetUserLogsResponse{
		UserLogs: protoLogs,
	}, nil
}

// convertUserModelToProto 将用户模型转换为proto格式
func (s *UserGRPCServer) convertUserModelToProto(userModel *model.User) *user.User {
	return &user.User{
		Id:               userModel.ID,
		Name:             userModel.Name,
		Email:            userModel.Email,
		Phone:            userModel.Phone,
		Avatar:           userModel.Avatar,
		Nickname:         userModel.Nickname,
		MemberShipLevel:  userModel.MemberShipLevel,
		MembershipExpire: timestamppb.New(userModel.MembershipExpire),
		Balance:          userModel.Balance,
		CreatedAt:        timestamppb.New(userModel.CreatedAt),
		UpdatedAt:        timestamppb.New(userModel.UpdatedAt),
	}
}

// convertUserSettingsModelToProto 将用户设置模型转换为proto格式
func (s *UserGRPCServer) convertUserSettingsModelToProto(userSettings *model.UserSettings) *user.UserSettings {
	return &user.UserSettings{
		UserId:             userSettings.UserID,
		LanguagePreference: userSettings.LanguagePreference,
		CreatedAt:          timestamppb.New(userSettings.CreatedAt),
		UpdatedAt:          timestamppb.New(userSettings.UpdatedAt),
	}
}

// convertUserPreferencesModelToProto 将用户偏好模型转换为proto格式
func (s *UserGRPCServer) convertUserPreferencesModelToProto(userPreferences *model.UserPreferences) *user.UserPreferences {
	return &user.UserPreferences{
		UserId:    userPreferences.UserID,
		TechArea:  userPreferences.TechArea,
		CreatedAt: timestamppb.New(userPreferences.CreatedAt),
		UpdatedAt: timestamppb.New(userPreferences.UpdatedAt),
	}
}

// convertUserLogsModelToProto 将用户日志模型转换为proto格式
func (s *UserGRPCServer) convertUserLogsModelToProto(userLogs *model.UserLogs) *user.UserLogs {
	return &user.UserLogs{
		Id:        userLogs.ID,
		UserId:    userLogs.UserID,
		Action:    userLogs.Action,
		IpAddress: userLogs.IPAddress,
		CreatedAt: timestamppb.New(userLogs.CreatedAt),
		UpdatedAt: timestamppb.New(userLogs.UpdatedAt),
	}
}
