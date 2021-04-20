package entity

import "database/sql"

type UserVariable struct {
	Name  string `gorm:"primaryKey"`
	Value sql.NullString
}
