package database

import "arkhive.dev/launcher/internal/console"

func (databaseEngine *DatabaseEngine) GetConsolePluginsByConsole(console *console.Console) (entity []console.ConsolePlugin, err error) {
	err = databaseEngine.database.Model(console).Association("ConsolePlugins").Find(&entity)
	return
}
