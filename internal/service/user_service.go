package service

import (
	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"github.com/cheel98/flashcard-backend/pkg/jwt"
	"go.uber.org/zap"
)

// UserService 用户服务接口
type UserService interface {
	Login(email, passwordHash string) (*jwt.TokenPair, error)
	RefreshToken(refreshToken string) (string, error)
	Logout(userID string) error
	GetUserByID(userID string) (*model.User, error)
	GetUserSettings(userID string) (*model.UserSettings, error)
	GetUserPreferences(userID string) (*model.UserPreferences, error)
	GetUserLogs(userID string, limit, offset int) ([]*model.UserLogs, error)
}

// userService 用户服务实现
type userService struct {
	userRepo   repository.UserRepository
	jwtManager *jwt.JWTManager
	logger     *zap.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, jwtManager *jwt.JWTManager, logger *zap.Logger) UserService {
	return &userService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

// Login 用户登录
func (s *userService) Login(email, passwordHash string) (*jwt.TokenPair, error) {
	s.logger.Info("用户尝试登录", zap.String("email", email))

	user, err := s.userRepo.Login(email, passwordHash)
	if err != nil {
		s.logger.Error("用户登录失败", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	// 生成token对
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		s.logger.Error("生成token失败", zap.String("userID", user.ID), zap.Error(err))
		return nil, err
	}

	// 保存refresh token到数据库
	err = s.userRepo.SaveRefreshToken(user.ID, tokenPair.RefreshToken)
	if err != nil {
		s.logger.Error("保存refresh token失败", zap.String("userID", user.ID), zap.Error(err))
		return nil, err
	}

	s.logger.Info("用户登录成功", zap.String("email", email), zap.String("userID", user.ID))
	return tokenPair, nil
}

// RefreshToken 刷新访问令牌
func (s *userService) RefreshToken(refreshToken string) (string, error) {
	s.logger.Debug("刷新访问令牌")

	// 验证refresh token
	user, err := s.userRepo.GetUserByRefreshToken(refreshToken)
	if err != nil {
		s.logger.Error("无效的refresh token", zap.Error(err))
		return "", err
	}

	// 生成新的access token
	accessToken, err := s.jwtManager.RefreshAccessToken(refreshToken)
	if err != nil {
		s.logger.Error("刷新access token失败", zap.String("userID", user.ID), zap.Error(err))
		return "", err
	}

	s.logger.Info("access token刷新成功", zap.String("userID", user.ID))
	return accessToken, nil
}

// Logout 用户登出
func (s *userService) Logout(userID string) error {
	s.logger.Info("用户登出", zap.String("userID", userID))

	// 清除refresh token
	err := s.userRepo.ClearRefreshToken(userID)
	if err != nil {
		s.logger.Error("清除refresh token失败", zap.String("userID", userID), zap.Error(err))
		return err
	}

	s.logger.Info("用户登出成功", zap.String("userID", userID))
	return nil
}

// GetUserByID 根据用户ID获取用户信息
func (s *userService) GetUserByID(userID string) (*model.User, error) {
	s.logger.Debug("获取用户信息", zap.String("userID", userID))

	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}

	return user, nil
}

// GetUserSettings 获取用户设置
func (s *userService) GetUserSettings(userID string) (*model.UserSettings, error) {
	s.logger.Debug("获取用户设置", zap.String("userID", userID))

	settings, err := s.userRepo.GetUserSettings(userID)
	if err != nil {
		s.logger.Error("获取用户设置失败", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}

	return settings, nil
}

// GetUserPreferences 获取用户偏好设置
func (s *userService) GetUserPreferences(userID string) (*model.UserPreferences, error) {
	s.logger.Debug("获取用户偏好设置", zap.String("userID", userID))

	preferences, err := s.userRepo.GetUserPreferences(userID)
	if err != nil {
		s.logger.Error("获取用户偏好设置失败", zap.String("userID", userID), zap.Error(err))
		return nil, err
	}

	return preferences, nil
}

// GetUserLogs 获取用户操作日志
func (s *userService) GetUserLogs(userID string, limit, offset int) ([]*model.UserLogs, error) {
	s.logger.Debug("获取用户操作日志",
		zap.String("userID", userID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	logs, err := s.userRepo.GetUserLogs(userID, limit, offset)
	if err != nil {
		s.logger.Error("获取用户操作日志失败",
			zap.String("userID", userID),
			zap.Error(err))
		return nil, err
	}

	return logs, nil
}
