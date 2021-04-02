package models

type ToolFileType struct {
	ToolID string `gorm:"not null"`
	Tool   Tool
	Type   string `gorm:"not null"`
}
