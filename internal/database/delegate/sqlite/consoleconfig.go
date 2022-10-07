package sqlite

type ConsoleConfig struct {
	ConsoleID string `gorm:"not null"`
	Name      string `gorm:"not null"`
	Value     string `gorm:"not null"`
	Level     string `gorm:"not null"` // ToDo: Handle enum
}

func (d *SQLiteDelegate) GetConsoleConfigs() (entity []ConsoleConfig, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
