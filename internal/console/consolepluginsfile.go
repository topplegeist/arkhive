package console

import "database/sql"

type ConsolePluginsFile struct {
	ConsolePluginID uint `gorm:"not null"`
	ConsolePlugin   ConsolePlugin
	Url             string `gorm:"not null"`
	Destination     sql.NullString
	CollectionPath  sql.NullString
}
