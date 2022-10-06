package importer

import (
	"errors"
)

type Tool struct {
	Slug           string
	Url            string
	CollectionPath *string
	Destination    *string
	Types           []string
}

func PlainDatabaseToTool(slug string, json interface{}) (tool Tool, err error) {
	var (
		entityObject map[string]interface{}
		ok           bool
	)
	if entityObject, ok = json.(map[string]interface{}); !ok {
		err = errors.New("the console JSON is not an object")
		return
	}

	if tool, err = ToolFromJSON(slug, entityObject); err != nil {
		return
	}

	if toolFileTypesObject, ok := entityObject["file_types"].([]interface{}); ok {
		for _, toolFileTypeObject := range toolFileTypesObject {
			var toolFileType string
			if toolFileType, ok = toolFileTypeObject.(string); ok {
				tool.Types = append(tool.Types, toolFileType)
			}
		}
	}
	return
}

func ToolFromJSON(slug string, json map[string]interface{}) (instance Tool, err error) {
	var collectionPath *string
	if collectionPathObject, ok := json["collection_path"].(string); ok {
		collectionPath = &collectionPathObject
	}
	var destination *string
	if destinationObject, ok := json["destination"].(string); ok {
		destination = &destinationObject
	}
	instance = Tool{
		slug,
		json["url"].(string),
		collectionPath,
		destination,
		[]string{},
	}
	return
}
