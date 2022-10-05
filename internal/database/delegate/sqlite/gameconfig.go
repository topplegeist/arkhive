package sqlite

import (
	"database/sql"

	"arkhive.dev/launcher/internal/database/importer"
)

type GameConfig struct {
	GameID string `gorm:"not null"`
	Name   string `gorm:"not null"`
	Value  string `gorm:"not null"`
}

func (d *SQLiteDelegate) storeImportedGameConfig(slug string, importedEntity importer.GameConfig) (err error) {
	entity := GameConfig{
		slug,
		importedEntity.Name,
		importedEntity.Value,
	}

	if entityCreationTransaction := d.database.Create(&entity); entityCreationTransaction.Error != nil {
		return entityCreationTransaction.Error
	}

	return
}
