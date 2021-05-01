package database

import "arkhive.dev/launcher/internal/entity"

func (databaseEngine *DatabaseEngine) GetConsolePluginsFilesByConsolePlugin(consolePlugin *entity.ConsolePlugin) (entity []entity.ConsolePluginsFile, err error) {
	err = databaseEngine.database.Model(consolePlugin).Association("ConsolePluginsFiles").Find(&entity)
	return
}
