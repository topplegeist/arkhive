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

func (d *SQLite) storeImportedConsole(importedEntity importer.Console) (err error) {
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

	if err = d.create(&entity); err != nil {
		return
	}

	for _, plugin := range importedEntity.Plugins {
		if err = d.storeImportedConsolePlugin(entity.Slug, plugin); err != nil {
			return
		}
	}

	for _, fileType := range importedEntity.FileTypes {
		if err = d.storeImportedConsoleFileType(entity.Slug, fileType); err != nil {
			return
		}
	}

	for _, config := range importedEntity.Configs {
		if err = d.storeImportedConsoleConfig(entity.Slug, config); err != nil {
			return
		}
	}

	for _, language := range importedEntity.Languages {
		if err = d.storeImportedConsoleLanguage(entity.Slug, language); err != nil {
			return
		}
	}

	return
}

func (d *SQLite) GetConsoles() (entity []Console, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (d *SQLite) GetConsoleByConsolePlugin(consolePlugin *ConsolePlugin) (model Console, err error) {
	err = d.database.Model(consolePlugin).Association("Console").Find(&model)
	return
}
