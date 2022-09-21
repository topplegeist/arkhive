package database

import "arkhive.dev/launcher/internal/console"

func (databaseEngine *DatabaseEngine) GetConsoles() (entity []console.Console, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (databaseEngine *DatabaseEngine) GetConsoleByConsolePlugin(consolePlugin *console.ConsolePlugin) (model console.Console, err error) {
	err = databaseEngine.database.Model(consolePlugin).Association("Console").Find(&model)
	return
}
