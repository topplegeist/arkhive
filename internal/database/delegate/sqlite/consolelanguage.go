package sqlite

type ConsoleLanguage struct {
	ConsoleID string `gorm:"not null"`
	Tag       uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
}

func (d *SQLiteDelegate) GetConsoleLanguages() (entity []ConsoleLanguage, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
