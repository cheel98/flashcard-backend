package model

import "time"

type Model struct {
	CreatedAt time.Time `gorm:"<-:create;column:create_at;type:datetime:default:now()" json:"created_at"`
	UpdateAt  time.Time `gorm:"column:update_at;type:datetime;default:now()" json:"update_at"`
	DeletedAt time.Time `gorm:"column:deleted_at;type:datetime;" json:"deleted_at"`
}
