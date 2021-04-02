package models

type ConsolePlugin struct {
	Id        uint   `gorm:"primaryKey"`
	ConsoleID string `gorm:"not null"`
	Console   Console
	Type      string `gorm:"not null"`
}
