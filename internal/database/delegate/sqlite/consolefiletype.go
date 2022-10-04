package sqlite

type ConsoleFileType struct {
	ConsoleID string `gorm:"not null"`
	Console   Console
	FileType  string `gorm:"not null"`
	Action    string `gorm:"not null"` // ToDo: Handle enum
}
