package sqlite

type ToolFilesType struct {
	ToolID string `gorm:"not null"`
	Type   string `gorm:"not null"`
}

func (d *SQLiteDelegate) storeImportedToolFilesType(slug string, toolType string) (err error) {
	entity := ToolFilesType{
		slug,
		toolType,
	}

	if err = d.create(&entity); err != nil {
		return
	}
	return
}

func (databaseEngine *SQLiteDelegate) GetToolFileTypes() (entity []ToolFilesType, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}
