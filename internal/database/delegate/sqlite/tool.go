package sqlite

import "database/sql"

type Tool struct {
	Slug           string `gorm:"primaryKey"`
	Url            string `gorm:"not null"`
	CollectionPath sql.NullString
	Destination    sql.NullString
}

func (databaseEngine *SQLiteDelegate) GetTools() (entity []Tool, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
