package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID               string    `gorm:"column:id;primary_key;type:varchar(255)" json:"id"`
	Name             string    `gorm:"column:name;uniqueIndex;type:varchar(20)" json:"name"`
	Email            string    `gorm:"column:email;uniqueIndex;type:varchar(255)" json:"email"`
	Phone            string    `gorm:"column:phone;uniqueIndex;type:varchar(11)" json:"phone"`
	PasswordHash     string    `gorm:"column:password_hash;type:varchar(255)" json:"-"`
	RefreshToken     string    `gorm:"column:refresh_token;type:text" json:"-"` // 刷新令牌
	Avatar           string    `gorm:"column:avatar;type:varchar(255)" json:"-"`
	Nickname         string    `gorm:"column:nickname" json:"-"`
	MemberShipLevel  uint64    `gorm:"column:member_ship_level;type:bigint" json:"-"`
	MembershipExpire time.Time `gorm:"column:membership_expire;type:timestamptz" json:"-"` // 会员到期时间
	Balance          uint64    `gorm:"column:balance;type:bigint" json:"-"`
	Model

	// 关联表
	UserSettings      UserSettings    `gorm:"foreignKey:UserID" json:"user_settings"`
	UserPreferences   UserPreferences `gorm:"foreignKey:UserID" json:"user_preferences"`
	UserLogs          []UserLogs      `gorm:"foreignKey:UserID" json:"user_logs"`
	PaymentRecords    []PaymentRecord `gorm:"foreignKey:UserID" json:"payment_records"`
	TranslationRecord []Favorite      `gorm:"foreignKey:UserID;references:ID" json:"favorites"`
}
type UserSettings struct {
	UserID             string `gorm:"column:user_id;uniqueIndex" json:"user_id"`
	LanguagePreference string `gorm:"column:language_preference" json:"language"`
	Model
}

type UserPreferences struct {
	UserID   string `gorm:"column:user_id;uniqueIndex" json:"user_id"`
	TechArea string `gorm:"column:tech_area" json:"tech_area"`
	Model
}
type UserLogs struct {
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint" json:"id"`
	UserID    string `gorm:"column:user_id;type:varchar(255)" json:"user_id"`
	Action    string `gorm:"column:action;type:varchar(255)" json:"action"`
	IPAddress string `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	Model
}

func (User) TableName() string {
	return "user"
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		uid, err := uuid.NewV7()
		if err != nil {
			return nil
		}
		u.ID = uid.String()
	}
	return nil
}

func (UserSettings) TableName() string {
	return "user_setting"
}
func (UserPreferences) TableName() string {
	return "user_preference"
}
func (UserLogs) TableName() string {
	return "user_log"
}
