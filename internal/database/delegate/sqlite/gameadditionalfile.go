package sqlite

import "arkhive.dev/launcher/internal/database/importer"

type GameAdditionalFile struct {
	GameID string `gorm:"not null"`
	Name   string `gorm:"not null"`
	Data   []byte `gorm:"not null"`
}

func (d *SQLiteDelegate) storeImportedGameAdditionalFile(slug string, importedEntity importer.GameAdditionalFile) (err error) {
	entity := GameAdditionalFile{
		slug,
		importedEntity.Name,
		importedEntity.Data,
	}

	if entityCreationTransaction := d.database.Create(&entity); entityCreationTransaction.Error != nil {
		return entityCreationTransaction.Error
	}

	return
}

func (d *SQLiteDelegate) GetGameAdditionalFiles() (entity []GameAdditionalFile, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
