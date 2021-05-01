package database

import "arkhive.dev/launcher/internal/entity"

func (databaseEngine *DatabaseEngine) GetConsolePluginsByConsole(console *entity.Console) (entity []entity.ConsolePlugin, err error) {
	err = databaseEngine.database.Model(console).Association("ConsolePlugins").Find(&entity)
	return
}
