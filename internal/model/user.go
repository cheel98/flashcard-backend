package model

import "time"

type User struct {
	ID               string    `gorm:"column:id;primary_key;type:varchar(255)" json:"id"`
	Name             string    `gorm:"column:name;uniqueIndex;type:varchar(20)" json:"name"`
	Email            string    `gorm:"column:email;uniqueIndex;type:varchar(255)" json:"email"`
	Phone            string    `gorm:"column:phone;uniqueIndex;type:varchar(11)" json:"phone"`
	PasswordHash     string    `gorm:"column:password_hash;type:varchar(255)" json:"-"`
	Avatar           string    `gorm:"column:avatar;type:varchar(255)" json:"-"`
	Nickname         string    `gorm:"column:nickname" json:"-"`
	MemberShipLevel  uint64    `gorm:"column:member_ship_level;type:bigint unsigned" json:"-"`
	MembershipExpire time.Time `gorm:"column:membership_expire;type:datetime" json:"-"` // 会员到期时间
	Balance          uint64    `gorm:"column:balance;type:bigint unsigned" json:"-"`
	Model

	// 关联表
	UserSettings      UserSettings    `gorm:"foreignKey:UserID" json:"user_settings"`
	UserPreferences   UserPreferences `gorm:"foreignKey:UserID" json:"user_preferences"`
	UserLogs          UserLogs        `gorm:"foreignKey:UserID" json:"user_logs"`
	PaymentRecords    []PaymentRecord `gorm:"foreignKey:UserID" json:"payment_records"`
	TranslationRecord []Favorite      `gorm:"column:favorites;foreignKey:UserID;references:ID" json:"favorites"`
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
	ID        uint64 `gorm:"primaryKey;autoIncrement;type:bigint unsigned" json:"id"`
	UserID    string `gorm:"column:user_id;type:varchar(255)" json:"user_id"`
	Action    string `gorm:"column:action;type:varchar(255)" json:"action"`
	IPAddress string `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	Model
}
