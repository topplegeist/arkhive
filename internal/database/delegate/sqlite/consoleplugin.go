package sqlite

import "arkhive.dev/launcher/internal/database/importer"

type ConsolePlugin struct {
	Id        uint   `gorm:"primaryKey;autoIncrement"`
	ConsoleID string `gorm:"not null"`
	Type      string `gorm:"not null"`
}

func (d *SQLite) storeImportedConsolePlugin(consoleId string, importedEntity importer.ConsolePlugin) (err error) {
	entity := ConsolePlugin{
		ConsoleID: consoleId,
		Type:      importedEntity.Type,
	}

	if err = d.create(&entity); err != nil {
		return
	}

	for _, file := range importedEntity.Files {
		if err = d.storeImportedConsolePluginsFile(entity.Id, file); err != nil {
			return
		}
	}

	return
}

func (d *SQLite) GetConsolePlugins() (entity []ConsolePlugin, err error) {
	if result := d.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (d *SQLite) GetConsolePluginsByConsole(console *Console) (entity []ConsolePlugin, err error) {
	err = d.database.Model(console).Association("ConsolePlugins").Find(&entity)
	return
}
