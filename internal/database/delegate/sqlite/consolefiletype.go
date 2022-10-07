package sqlite

type ConsoleFileType struct {
	ConsoleID string `gorm:"not null"`
	FileType  string `gorm:"not null"`
	Action    string `gorm:"not null"` // ToDo: Handle enum
}

func (d *SQLiteDelegate) GetConsoleFileTypes() (entity []ConsoleFileType, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
