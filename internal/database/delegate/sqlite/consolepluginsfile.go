package sqlite

import (
	"database/sql"
)

type ConsolePluginsFile struct {
	ConsolePluginID uint `gorm:"not null"`
	ConsolePlugin   ConsolePlugin
	Url             string `gorm:"not null"`
	Destination     sql.NullString
	CollectionPath  sql.NullString
}

func (d *SQLiteDelegate) GetConsolePluginsFilesByConsolePlugin(consolePlugin *ConsolePlugin) (entity []ConsolePluginsFile, err error) {
	err = d.database.Model(consolePlugin).Association("ConsolePluginsFiles").Find(&entity)
	return
}
