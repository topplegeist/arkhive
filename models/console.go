package models

import (
	"database/sql"
)

type Console struct {
	Slug                 string `gorm:"primaryKey"`
	CoreLocation         string `gorm:"not null"`
	Name                 string `gorm:"not null"`
	SingleFile           bool   `gorm:"not null"`
	LanguageVariableName sql.NullString
	IsEmbedded           bool `gorm:"not null"`
	ConsolePlugins       []ConsolePlugin
}

func ConsoleFromJSON(consoleSlug string, json interface{}) (instance *Console, err error) {
	languageVariableName := sql.NullString{String: "", Valid: false}
	if languageObject, ok := json.(map[string]interface{})["language"]; ok {
		if languageVariableNameObject, ok := languageObject.(map[string]interface{})["variable_name"]; ok {
			languageVariableName.String = languageVariableNameObject.(string)
			languageVariableName.Valid = true
		}
	}
	instance = &Console{
		consoleSlug,
		json.(map[string]interface{})["core_location"].(string),
		json.(map[string]interface{})["name"].(string),
		json.(map[string]interface{})["single_file"].(bool),
		languageVariableName,
		false,
		[]ConsolePlugin{},
	}
	return
}
