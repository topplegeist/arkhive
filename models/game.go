package models

import (
	"database/sql"
	"time"
)

type Game struct {
	Slug            string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	ConsoleID       string `gorm:"not null"`
	Console         Console
	BackgroundColor string `gorm:"not null"`
	BackgroundImage sql.NullString
	Logo            sql.NullString
	Executable      sql.NullString
	InsertionDate   time.Time `gorm:"autoCreateTime;not null"`
}
