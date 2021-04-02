package models

type ConsoleConfig struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	Name      string `gorm:"not null"`
	Value     string `gorm:"not null"`
	Level     string `gorm:"not null"`
}
