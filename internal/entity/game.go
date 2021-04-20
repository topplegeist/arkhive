package entity

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

func GameFromJSON(gameSlug string, console *Console, json interface{}) (instance *Game, err error) {
	backgroundImage := sql.NullString{String: "", Valid: false}
	if backgroundImageObject, ok := json.(map[string]interface{})["background_image"].(string); ok {
		backgroundImage.String = backgroundImageObject
		backgroundImage.Valid = true
	}
	logo := sql.NullString{String: "", Valid: false}
	if logoObject, ok := json.(map[string]interface{})["logo"].(string); ok {
		logo.String = logoObject
		logo.Valid = true
	}
	executable := sql.NullString{String: "", Valid: false}
	if executableObject, ok := json.(map[string]interface{})["executable"].(string); ok {
		executable.String = executableObject
		executable.Valid = true
	}
	instance = &Game{
		gameSlug,
		json.(map[string]interface{})["name"].(string),
		"",
		*console,
		json.(map[string]interface{})["background_color"].(string),
		backgroundImage,
		logo,
		executable,
		time.Now(),
	}
	return
}
