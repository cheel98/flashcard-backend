package repository

import (
	"errors"
	"github.com/cheel98/flashcard-backend/internal/model"
	"gorm.io/gorm"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(*model.User) (*model.User, error)
	// Login 用户登录验证
	Login(email, passwordHash string) (*model.User, error)
	// GetUserByID 根据ID获取用户基本信息
	GetUserByID(userID string) (*model.User, error)
	// GetUserSettings 获取用户设置
	GetUserSettings(userID string) (*model.UserSettings, error)
	// GetUserPreferences 获取用户个人喜好
	GetUserPreferences(userID string) (*model.UserPreferences, error)
	// GetUserLogs 获取用户操作日志
	GetUserLogs(userID string, limit, offset int) ([]*model.UserLogs, error)
	// SaveRefreshToken 保存刷新令牌
	SaveRefreshToken(userID, refreshToken string) error
	// GetUserByRefreshToken 根据刷新令牌获取用户
	GetUserByRefreshToken(refreshToken string) (*model.User, error)
	// ClearRefreshToken 清除刷新令牌
	ClearRefreshToken(userID string) error
}

// userRepository 用户仓储实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户仓储实例
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}
func (r *userRepository) Create(user *model.User) (*model.User, error) {
	err := r.db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Login 用户登录验证
func (r *userRepository) Login(email, passwordHash string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ? AND password_hash = ?", email, passwordHash).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户基本信息（不使用关联查询）
func (r *userRepository) GetUserByID(userID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserSettings 获取用户设置（不使用关联查询）
func (r *userRepository) GetUserSettings(userID string) (*model.UserSettings, error) {
	var settings model.UserSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户设置不存在")
		}
		return nil, err
	}
	return &settings, nil
}

// GetUserPreferences 获取用户个人喜好（不使用关联查询）
func (r *userRepository) GetUserPreferences(userID string) (*model.UserPreferences, error) {
	var preferences model.UserPreferences
	err := r.db.Where("user_id = ?", userID).First(&preferences).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户喜好设置不存在")
		}
		return nil, err
	}
	return &preferences, nil
}

// GetUserLogs 获取用户操作日志（不使用关联查询）
func (r *userRepository) GetUserLogs(userID string, limit, offset int) ([]*model.UserLogs, error) {
	var logs []*model.UserLogs
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	return logs, nil
}

// SaveRefreshToken 保存刷新令牌
func (r *userRepository) SaveRefreshToken(userID, refreshToken string) error {
	err := r.db.Model(&model.User{}).Where("id = ?", userID).Update("refresh_token", refreshToken).Error
	if err != nil {
		return err
	}
	return nil
}

// GetUserByRefreshToken 根据刷新令牌获取用户
func (r *userRepository) GetUserByRefreshToken(refreshToken string) (*model.User, error) {
	var user model.User
	err := r.db.Where("refresh_token = ? AND refresh_token != ''", refreshToken).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("无效的刷新令牌")
		}
		return nil, err
	}
	return &user, nil
}

// ClearRefreshToken 清除刷新令牌
func (r *userRepository) ClearRefreshToken(userID string) error {
	err := r.db.Model(&model.User{}).Where("id = ?", userID).Update("refresh_token", "").Error
	if err != nil {
		return err
	}
	return nil
}
