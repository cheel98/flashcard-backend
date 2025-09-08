package model

import "time"

// 充值记录表
type PaymentRecord struct {
	ID            string     `gorm:"primaryKey;type:varchar(255)"`
	UserID        string     `gorm:"not null;index;type:varchar(255)"`
	Amount        float64    `gorm:"type:decimal(10,2);not null"` // 金额
	Currency      string     `gorm:"type:varchar(10);default:'CNY'"`
	PaymentMethod string     `gorm:"type:varchar(50)"`                   // 微信、支付宝、信用卡等
	Status        string     `gorm:"type:varchar(20);default:'pending'"` // pending, completed, failed
	TransactionID string     `gorm:"type:varchar(100)"`                  // 第三方支付流水号
	CreatedAt     time.Time  `gorm:"autoCreateTime;type:datetime"`
	CompletedAt   *time.Time `gorm:"type:datetime"`
}

// 会员权益表
type MembershipBenefit struct {
	ID               string  `gorm:"primaryKey;type:bigint unsigned"`
	Level            string  `gorm:"type:varchar(20);not null"` // silver, gold, platinum
	MaxFlashcards    int     `gorm:"default:1000;type:int"`     // 每月可生成闪记卡数量
	TranslationLimit int     `gorm:"default:100;type:int"`      // 每日翻译次数
	Price            float64 `gorm:"type:decimal(10,2)"`        // 价格
}
