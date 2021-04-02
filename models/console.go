package models

import "database/sql"

type Console struct {
	Slug                 string `gorm:"primaryKey"`
	CoreLocation         string `gorm:"not null"`
	Name                 string `gorm:"not null"`
	SingleFile           bool   `gorm:"not null"`
	LanguageVariableName sql.NullString
	IsEmbedded           bool `gorm:"not null"`
}
