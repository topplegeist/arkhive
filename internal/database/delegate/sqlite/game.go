package sqlite

import (
	"database/sql"
	"time"

	"arkhive.dev/launcher/internal/database/importer"
)

type Game struct {
	Slug            string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	ConsoleID       string `gorm:"not null"`
	BackgroundColor string `gorm:"not null"`
	BackgroundImage sql.NullString
	Logo            sql.NullString
	Executable      sql.NullString
	InsertionDate   time.Time `gorm:"autoCreateTime;not null"`
}

func (d *SQLiteDelegate) storeImportedGame(importedEntity importer.Game) (err error) {
	backgroundImage := sql.NullString{}
	if importedEntity.BackgroundImage != nil {
		backgroundImage.Valid = true
		backgroundImage.String = *importedEntity.BackgroundImage
	}
	logo := sql.NullString{}
	if importedEntity.Logo != nil {
		logo.Valid = true
		logo.String = *importedEntity.Logo
	}
	executable := sql.NullString{}
	if importedEntity.Executable != nil {
		executable.Valid = true
		executable.String = *importedEntity.Executable
	}
	entity := Game{
		importedEntity.Slug,
		importedEntity.Name,
		importedEntity.ConsoleSlug,
		importedEntity.BackgroundColor,
		backgroundImage,
		logo,
		executable,
		time.Now(),
	}

	if entityCreationTransaction := d.database.Create(&entity); entityCreationTransaction.Error != nil {
		return entityCreationTransaction.Error
	}

	for _, disk := range importedEntity.Disks {
		if err = d.storeImportedGameDisk(entity.Slug, disk); err != nil {
			return
		}
	}
	for _, config := range importedEntity.Configs {
		if err = d.storeImportedGameConfig(entity.Slug, config); err != nil {
			return
		}
	}
	for _, additionalFile := range importedEntity.AdditionalFiles {
		if err = d.storeImportedGameAdditionalFile(entity.Slug, additionalFile); err != nil {
			return
		}
	}
	return
}

func (d *SQLiteDelegate) GetGames() (entity []Game, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
