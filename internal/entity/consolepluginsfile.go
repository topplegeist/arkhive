package entity

import "database/sql"

type ConsolePluginsFile struct {
	ConsolePluginID uint `gorm:"not null"`
	ConsolePlugin   ConsolePlugin
	Url             string `gorm:"not null"`
	Destination     sql.NullString
	CollectionPath  sql.NullString
}

func ConsolePluginsFileFromJSON(consolePlugin *ConsolePlugin, jsonCollectionPath interface{},
	jsonDestination interface{}, jsonFile interface{}) (instance *ConsolePluginsFile, err error) {
	destination := sql.NullString{String: "", Valid: false}
	if destinationObject, ok := jsonDestination.(string); ok {
		destination.String = destinationObject
		destination.Valid = true
	}
	collectionPath := sql.NullString{String: "", Valid: false}
	if collectionPathObject, ok := jsonCollectionPath.(string); ok {
		collectionPath.String = collectionPathObject
		collectionPath.Valid = true
	}
	instance = &ConsolePluginsFile{
		0,
		*consolePlugin,
		jsonFile.(string),
		destination,
		collectionPath,
	}
	return
}
