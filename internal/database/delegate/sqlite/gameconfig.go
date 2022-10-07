package sqlite

import (
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

	if err = d.create(&entity); err != nil {
		return
	}

	return
}

func (d *SQLiteDelegate) GetGameConfigs() (entity []GameConfig, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
