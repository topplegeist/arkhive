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

	if entityCreationTransaction := d.database.Create(&entity); entityCreationTransaction.Error != nil {
		return entityCreationTransaction.Error
	}
	return
}
