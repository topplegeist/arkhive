package tool

import "database/sql"

type Tool struct {
	Slug           string `gorm:"primaryKey"`
	Url            string `gorm:"not null"`
	CollectionPath sql.NullString
	Destination    sql.NullString
}

func ToolFromJSON(slug string, json interface{}) (instance *Tool, err error) {
	collectionPath := sql.NullString{String: "", Valid: false}
	if collectionPathObject, ok := json.(map[string]interface{})["collection_path"].(string); ok {
		collectionPath.String = collectionPathObject
		collectionPath.Valid = true
	}
	destination := sql.NullString{String: "", Valid: false}
	if destinationObject, ok := json.(map[string]interface{})["destination"].(string); ok {
		destination.String = destinationObject
		destination.Valid = true
	}
	instance = &Tool{
		slug,
		json.(map[string]interface{})["url"].(string),
		collectionPath,
		destination,
	}
	return
}
