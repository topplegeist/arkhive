package sqlite

import (
	"database/sql"

	"arkhive.dev/launcher/internal/database/importer"
)

type ConsolePluginsFile struct {
	ConsolePluginID uint   `gorm:"not null"`
	Url             string `gorm:"not null"`
	Destination     sql.NullString
	CollectionPath  sql.NullString
}

func (d *SQLiteDelegate) storeImportedPluginsFile(consolePluginId uint, importedEntity importer.ConsolePluginsFile) (err error) {
	destination := sql.NullString{}
	if importedEntity.Destination != nil {
		destination.Valid = true
		destination.String = *importedEntity.Destination
	}
	collectionPath := sql.NullString{}
	if importedEntity.CollectionPath != nil {
		collectionPath.Valid = true
		collectionPath.String = *importedEntity.CollectionPath
	}
	entity := ConsolePluginsFile{
		ConsolePluginID: consolePluginId,
		Url:             importedEntity.Url,
		Destination:     destination,
		CollectionPath:  collectionPath,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	return
}

func (d *SQLiteDelegate) GetConsolePluginsFiles() (entity []ConsolePluginsFile, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (d *SQLiteDelegate) GetConsolePluginsFilesByConsolePlugin(consolePlugin *ConsolePlugin) (entity []ConsolePluginsFile, err error) {
	err = d.database.Model(consolePlugin).Association("ConsolePluginsFiles").Find(&entity)
	return
}
