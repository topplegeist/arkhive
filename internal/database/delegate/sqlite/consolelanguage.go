package sqlite

import "arkhive.dev/launcher/internal/database/importer"

type ConsoleLanguage struct {
	ConsoleID string `gorm:"not null"`
	Tag       uint   `gorm:"not null"`
	Name      string `gorm:"not null"`
}

func (d *SQLiteDelegate) storeImportedConsoleLanguage(consoleId string, importedEntity importer.ConsoleLanguage) (err error) {
	entity := ConsoleLanguage{
		ConsoleID: consoleId,
		Tag:       importedEntity.Tag,
		Name:      importedEntity.Name,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	return
}

func (d *SQLiteDelegate) GetConsoleLanguages() (entity []ConsoleLanguage, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
