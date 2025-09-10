package model

import "time"

type Model struct {
	CreatedAt time.Time `gorm:"<-:create;column:create_at;type:timestamptz;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:update_at;type:timestamptz;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt time.Time `gorm:"column:deleted_at;type:timestamptz" json:"deleted_at"`
}
