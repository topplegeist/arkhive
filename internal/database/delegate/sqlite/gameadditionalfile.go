package sqlite

import "arkhive.dev/launcher/internal/database/importer"

type GameAdditionalFile struct {
	GameID string `gorm:"not null"`
	Name   string `gorm:"not null"`
	Data   []byte `gorm:"not null"`
}

func (d *SQLite) storeImportedGameAdditionalFile(slug string, importedEntity importer.GameAdditionalFile) (err error) {
	entity := GameAdditionalFile{
		slug,
		importedEntity.Name,
		importedEntity.Data,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	return
}

func (d *SQLite) GetGameAdditionalFiles() (entity []GameAdditionalFile, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
