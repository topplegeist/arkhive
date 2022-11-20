package sqlite

import "arkhive.dev/launcher/internal/database/importer"

type ConsoleFileType struct {
	ConsoleID string `gorm:"not null"`
	FileType  string `gorm:"not null"`
	Action    string `gorm:"not null"`
}

func (d *SQLite) storeImportedConsoleFileType(consoleId string, importedEntity importer.ConsoleFileType) (err error) {
	entity := ConsoleFileType{
		ConsoleID: consoleId,
		FileType:  importedEntity.FileType,
		Action:    importedEntity.Action,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	return
}

func (d *SQLite) GetConsoleFileTypes() (entity []ConsoleFileType, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
