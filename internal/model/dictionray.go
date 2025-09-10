package model

type Dictionary struct {
	ID              uint64 `gorm:"primaryKey;autoIncrement;type:bigint" json:"id"`
	SourceLang      string `gorm:"type:varchar(10);not null;index:idx_unique_translation,unique" json:"source_lang"`
	TargetLang      string `gorm:"type:varchar(10);not null;index:idx_unique_translation,unique" json:"target_lang"`
	SourceText      string `gorm:"type:text;not null;index:idx_unique_translation,unique" json:"source_text"`
	TranslatedText  string `gorm:"type:text;not null" json:"translated_text"`
	PartOfSpeech    string `gorm:"type:varchar(20)" json:"part_of_speech"`
	IPA             string `gorm:"type:varchar(50)" json:"ipa"`
	ExampleSentence string `gorm:"type:text" json:"example_sentence"`
	Model

	// 关联表
	Audios   []DictionaryAudio    `gorm:"foreignKey:DictionaryID;constraint:OnDelete:CASCADE" json:"audios"`
	Metadata []DictionaryMetadata `gorm:"foreignKey:DictionaryID;constraint:OnDelete:CASCADE" json:"metadata"`
}

// DictionaryAudio 存储多发音文件
type DictionaryAudio struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement;type:bigint" json:"id"`
	DictionaryID uint64 `gorm:"not null;index;type:bigint" json:"dictionary_id"`
	AudioPath    string `gorm:"type:varchar(255);not null" json:"audio_path"`
	Accent       string `gorm:"type:varchar(50)" json:"accent"` // 美式/英式等
	Model
}

// DictionaryMetadata 存储额外信息，如例句、标签等
type DictionaryMetadata struct {
	ID           uint64 `gorm:"primaryKey;autoIncrement;type:bigint" json:"id"`
	DictionaryID uint64 `gorm:"not null;index;type:bigint" json:"dictionary_id"`
	Key          string `gorm:"type:varchar(50);not null" json:"key"` // 如 "synonym", "antonym"
	Value        string `gorm:"type:text;not null" json:"value"`      // 对应内容
	Model
}

func (Dictionary) TableName() string {
	return "dictionary"
}

func (DictionaryAudio) TableName() string {
	return "dictionary_audio"
}
func (DictionaryMetadata) TableName() string {
	return "dictionary_metadata"
}
