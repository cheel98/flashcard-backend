package service

import (
	"github.com/cheel98/flashcard-backend/internal/model"
	"github.com/cheel98/flashcard-backend/internal/repository"
	"go.uber.org/zap"
)

// UserService 用户服务接口
type UserService interface {
	Login(email, passwordHash string) (*model.User, error)
	GetUserByID(userID string) (*model.User, error)
	GetUserSettings(userID string) (*model.UserSettings, error)
	GetUserPreferences(userID string) (*model.UserPreferences, error)
	GetUserLogs(userID string, limit, offset int) ([]*model.UserLogs, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
	logger   *zap.Logger
}

// NewUserService 创建用户服务实例
func NewUserService(userRepo repository.UserRepository, logger *zap.Logger) UserService {
	return &userService{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Login 用户登录
func (s *userService) Login(email, passwordHash string) (*model.User, error) {
	s.logger.Info("用户尝试登录", zap.String("email", email))

	user, err := s.userRepo.Login(email, passwordHash)
	if err != nil {
		s.logger.Error("用户登录失败", zap.String("email", email), zap.Error(err))
		return nil, err
	}

	s.logger.Info("用户登录成功", zap.String("email", email), zap.String("userID", user.ID))
	return user, nil
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
