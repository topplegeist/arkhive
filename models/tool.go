package models

import "database/sql"

type Tool struct {
	Slug           string `gorm:"primaryKey"`
	Url            string `gorm:"not null"`
	CollectionPath sql.NullString
	Destination    sql.NullString
}
