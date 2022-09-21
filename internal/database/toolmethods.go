package database

import "arkhive.dev/launcher/internal/tool"

func (databaseEngine *DatabaseEngine) GetTools() (entity []tool.Tool, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
