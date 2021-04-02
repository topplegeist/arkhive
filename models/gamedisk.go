package models

import "database/sql"

type GameDisk struct {
	GameID         string `gorm:"not null"`
	Game           Game
	DiskNumber     uint   `gorm:"not null"`
	Url            string `gorm:"not null"`
	Image          sql.NullString
	CollectionPath sql.NullString
}
