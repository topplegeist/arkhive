package models

type ToolFilesType struct {
	ToolID string `gorm:"not null"`
	Tool   Tool
	Type   string `gorm:"not null"`
}

func ToolFilesTypeFromJSON(tool *Tool, json interface{}) (instance *ToolFilesType, err error) {
	instance = &ToolFilesType{
		"",
		*tool,
		json.(string),
	}
	return
}
