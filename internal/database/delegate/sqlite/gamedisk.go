package sqlite

import (
	"database/sql"

	"arkhive.dev/launcher/internal/database/importer"
)

type GameDisk struct {
	GameID         string `gorm:"not null"`
	DiskNumber     uint   `gorm:"not null"`
	Url            string `gorm:"not null"`
	Image          sql.NullString
	CollectionPath sql.NullString
}

func (d *SQLiteDelegate) storeImportedGameDisk(slug string, importedEntity importer.GameDisk) (err error) {
	image := sql.NullString{}
	if importedEntity.Image != nil {
		image.Valid = true
		image.String = *importedEntity.Image
	}
	collectionPath := sql.NullString{}
	if importedEntity.CollectionPath != nil {
		collectionPath.Valid = true
		collectionPath.String = *importedEntity.CollectionPath
	}
	entity := GameDisk{
		GameID:         slug,
		DiskNumber:     importedEntity.DiskNumber,
		Url:            importedEntity.Url,
		Image:          image,
		CollectionPath: collectionPath,
	}

	if entityCreationTransaction := d.database.Create(&entity); entityCreationTransaction.Error != nil {
		return entityCreationTransaction.Error
	}

	return
}
