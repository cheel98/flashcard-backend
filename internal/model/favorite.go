package model

// 收藏
type Favorite struct {
	ID              string        `gorm:"column:id;type:varchar(255)" json:"id"`
	UserID          string        `gorm:"column:user_id;type:varchar(255)" json:"user_id"`
	DictionaryID    uint64        `gorm:"column:dictionary_id;type:bigint" json:"dictionary_id"`
	MemoryDepth     uint64        `gorm:"column:memory_depth;type:bigint" json:"memory_depth"`
	FavoriteRecords []StudyRecord `gorm:"foreignKey:ID" json:"favorite_records"`
	Model
}

type StudyRecord struct {
	ID     string `gorm:"column:id;primary_key;type:varchar(255)" json:"id"`
	Result string `gorm:"column:result;type:varchar(20);check:result IN ('remembered','fuzzy','strange')" json:"result"` // 学习结果
	Remark string `gorm:"column:remark;type:text" json:"remark"`
	Model
}

func (Favorite) TableName() string {
	return "favorite"
}
func (StudyRecord) TableName() string {
	return "study_record"
}
