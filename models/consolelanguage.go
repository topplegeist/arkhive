package models

type ConsoleLanguage struct {
	ConsoleID string `gorm:"primaryKey"`
	Console   Console
	Id        uint   `gorm:"primaryKey;autoIncrement:false"`
	Name      string `gorm:"not null"`
}
