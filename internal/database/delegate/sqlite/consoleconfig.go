package sqlite

import "arkhive.dev/launcher/internal/database/importer"

type ConsoleConfig struct {
	ConsoleID string `gorm:"not null"`
	Name      string `gorm:"not null"`
	Value     string `gorm:"not null"`
	Level     string `gorm:"not null"`
}

func (d *SQLiteDelegate) storeImportedConsoleConfig(consoleId string, importedEntity importer.ConsoleConfig) (err error) {
	entity := ConsoleConfig{
		ConsoleID: consoleId,
		Name:      importedEntity.Name,
		Value:     importedEntity.Value,
		Level:     importedEntity.Level,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	return
}

func (d *SQLiteDelegate) GetConsoleConfigs() (entity []ConsoleConfig, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
