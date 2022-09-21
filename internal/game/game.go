package game

import (
	"database/sql"
	"time"

	"arkhive.dev/launcher/internal/console"
)

type Game struct {
	Slug            string `gorm:"primaryKey"`
	Name            string `gorm:"not null"`
	ConsoleID       string `gorm:"not null"`
	Console         console.Console
	BackgroundColor string `gorm:"not null"`
	BackgroundImage sql.NullString
	Logo            sql.NullString
	Executable      sql.NullString
	InsertionDate   time.Time `gorm:"autoCreateTime;not null"`
}

func GameFromJSON(gameSlug string, console *console.Console, json interface{}) (instance *Game, err error) {
	gameData := json.(map[string]interface{})
	backgroundImage := sql.NullString{String: "", Valid: false}
	if backgroundImageObject, ok := gameData["background_image"].(string); ok {
		backgroundImage.String = backgroundImageObject
		backgroundImage.Valid = true
	}
	logo := sql.NullString{String: "", Valid: false}
	if logoObject, ok := gameData["logo"].(string); ok {
		logo.String = logoObject
		logo.Valid = true
	}
	executable := sql.NullString{String: "", Valid: false}
	if executableObject, ok := gameData["executable"].(string); ok {
		executable.String = executableObject
		executable.Valid = true
	}
	instance = &Game{
		gameSlug,
		gameData["name"].(string),
		"",
		*console,
		gameData["background_color"].(string),
		backgroundImage,
		logo,
		executable,
		time.Now(),
	}
	return
}
