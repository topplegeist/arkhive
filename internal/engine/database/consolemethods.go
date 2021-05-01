package database

import "arkhive.dev/launcher/internal/entity"

func (databaseEngine *DatabaseEngine) GetConsoles() (entity []entity.Console, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (databaseEngine *DatabaseEngine) GetConsoleByConsolePlugin(consolePlugin *entity.ConsolePlugin) (model entity.Console, err error) {
	err = databaseEngine.database.Model(consolePlugin).Association("Console").Find(&model)
	return
}
