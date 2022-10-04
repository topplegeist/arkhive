package sqlite

type ConsolePlugin struct {
	Id                  uint   `gorm:"primaryKey"`
	ConsoleID           string `gorm:"not null"`
	Console             Console
	Type                string `gorm:"not null"`
	ConsolePluginsFiles []ConsolePluginsFile
}

func (d *SQLiteDelegate) GetConsolePluginsByConsole(console *Console) (entity []ConsolePlugin, err error) {
	err = d.database.Model(console).Association("ConsolePlugins").Find(&entity)
	return
}
