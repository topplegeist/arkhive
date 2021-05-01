package database

import "arkhive.dev/launcher/internal/entity"

func (databaseEngine *DatabaseEngine) GetTools() (entity []entity.Tool, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
