package database

import "arkhive.dev/launcher/internal/console"

func (databaseEngine *DatabaseEngine) GetConsolePluginsFilesByConsolePlugin(consolePlugin *console.ConsolePlugin) (entity []console.ConsolePluginsFile, err error) {
	err = databaseEngine.database.Model(consolePlugin).Association("ConsolePluginsFiles").Find(&entity)
	return
}
