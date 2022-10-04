package sqlite

type ConsoleLanguage struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	Tag       uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
}
