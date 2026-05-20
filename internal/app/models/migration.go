package models

type Migration struct {
	Version   string `gorm:"primaryKey;size:64" json:"version"`
	Name      string `gorm:"size:255" json:"name"`
	CreatedAt uint32 `json:"created_at"`
}

func (Migration) TableName() string {
	return "sys_migration"
}
