package models

import "encoding/base64"

type GameAdditionalFile struct {
	Name   string `gorm:"not null"`
	GameID string `gorm:"not null"`
	Game   Game
	Data   []byte `gorm:"not null"`
}

func GameAdditionalFileFromJSON(game *Game, json interface{}) (instance *GameAdditionalFile, err error) {
	var data []byte
	if data, err = base64.URLEncoding.DecodeString(json.(map[string]interface{})["base64"].(string)); err != nil {
		return
	}
	instance = &GameAdditionalFile{
		json.(map[string]interface{})["name"].(string),
		"",
		*game,
		data,
	}
	return
}
