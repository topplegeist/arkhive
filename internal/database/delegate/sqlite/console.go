package sqlite

import (
	"database/sql"

	"arkhive.dev/launcher/internal/database/importer"
)

type Console struct {
	Slug                 string `gorm:"primaryKey"`
	CoreLocation         string `gorm:"not null"`
	Name                 string `gorm:"not null"`
	SingleFile           bool   `gorm:"not null"`
	LanguageVariableName sql.NullString
	IsEmbedded           bool `gorm:"not null"`
}

func (d *SQLiteDelegate) storeImportedConsole(importedEntity importer.Console) (err error) {
	languageVariableName := sql.NullString{}
	if importedEntity.LanguageVariableName != nil {
		languageVariableName.Valid = true
		languageVariableName.String = *importedEntity.LanguageVariableName
	}
	entity := Console{
		importedEntity.Slug,
		importedEntity.CoreLocation,
		importedEntity.Name,
		importedEntity.SingleFile,
		languageVariableName,
		importedEntity.IsEmbedded,
	}

	if entityCreationTransaction := d.database.Create(&entity); entityCreationTransaction.Error != nil {
		return entityCreationTransaction.Error
	}

	for _, plugin := range importedEntity.Plugins {
		if err = d.storeImportedPlugin(entity.Slug, plugin); err != nil {
			return
		}
	}
	return
}

func (d *SQLiteDelegate) GetConsoles() (entity []Console, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (d *SQLiteDelegate) GetConsoleByConsolePlugin(consolePlugin *ConsolePlugin) (model Console, err error) {
	err = d.database.Model(consolePlugin).Association("Console").Find(&model)
	return
}
