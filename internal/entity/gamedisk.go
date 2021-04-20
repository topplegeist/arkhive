package entity

import (
	"database/sql"
)

type GameDisk struct {
	GameID         string `gorm:"not null"`
	Game           Game
	DiskNumber     uint   `gorm:"not null"`
	Url            string `gorm:"not null"`
	Image          sql.NullString
	CollectionPath sql.NullString
}

func GameDiskFromJSON(game *Game, diskNumber uint, jsonUrl interface{}, jsonDiskImage interface{}, jsonCollectionPath interface{}) (instance *GameDisk, err error) {
	image := sql.NullString{String: "", Valid: false}
	if imageObject, ok := jsonDiskImage.(string); ok {
		image.String = imageObject
		image.Valid = true
	}
	collectionPath := sql.NullString{String: "", Valid: false}
	if collectionPathObject, ok := jsonCollectionPath.(string); ok {
		collectionPath.String = collectionPathObject
		collectionPath.Valid = true
	}
	instance = &GameDisk{
		"",
		*game,
		diskNumber,
		jsonUrl.(string),
		image,
		collectionPath,
	}
	return
}
