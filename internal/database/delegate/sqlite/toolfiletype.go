package sqlite

type ToolFilesType struct {
	ToolID string `gorm:"not null"`
	Tool   Tool
	Type   string `gorm:"not null"`
}
