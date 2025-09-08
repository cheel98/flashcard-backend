package model

type User struct {
	ID           string `gorm:"column:id;primary_key" json:"id"`
	Name         string `gorm:"column:name;uniqueIndex" json:"name"`
	Email        string `gorm:"column:email;uniqueIndex" json:"email"`
	Phone        string `gorm:"column:phone;uniqueIndex" json:"phone"`
	PasswordHash string `gorm:"column:password_hash" json:"-"`
	Avatar       string `gorm:"column:avatar" json:"-"`
	Nickname     string `gorm:"column:nickname" json:"-"`

	Model
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
