package entity

import (
	"encoding/json"
	"errors"
	"fmt"
)

type GameConfig struct {
	GameID string `gorm:"not null"`
	Game   Game
	Name   string `gorm:"not null"`
	Value  string `gorm:"not null"`
}

func GameConfigFromJSON(game *Game, name string, jsonValue interface{}) (instance *GameConfig, err error) {
	var (
		value         string
		integerValue  int64
		floatingValue float64
	)
	if integerValue, err = jsonValue.(json.Number).Int64(); err == nil {
		value = fmt.Sprintf("%d", integerValue)
	} else if floatingValue, err = jsonValue.(json.Number).Float64(); err == nil {
		value = fmt.Sprintf("%f", floatingValue)
	} else {
		err = errors.New("wrong configuration variable value format")
		return
	}

	instance = &GameConfig{
		"",
		*game,
		name,
		value,
	}
	return
}
