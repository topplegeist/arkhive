package sqlite

import (
	"database/sql"

	"arkhive.dev/launcher/internal/database/importer"
)

type Tool struct {
	Slug           string `gorm:"primaryKey"`
	Url            string `gorm:"not null"`
	CollectionPath sql.NullString
	Destination    sql.NullString
}

func (d *SQLiteDelegate) storeImportedTool(importedEntity importer.Tool) (err error) {
	collectionPath := sql.NullString{}
	if importedEntity.CollectionPath != nil {
		collectionPath.Valid = true
		collectionPath.String = *importedEntity.CollectionPath
	}
	destination := sql.NullString{}
	if importedEntity.Destination != nil {
		destination.Valid = true
		destination.String = *importedEntity.Destination
	}
	entity := Tool{
		importedEntity.Slug,
		importedEntity.Url,
		collectionPath,
		destination,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	for _, toolType := range importedEntity.Types {
		if err = d.storeImportedToolFilesType(entity.Slug, toolType); err != nil {
			return
		}
	}
	return
}

func (databaseEngine *SQLiteDelegate) GetTools() (entity []Tool, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
