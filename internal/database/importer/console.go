package importer

import (
	"errors"
	"strconv"
)

var consoleConfigLevels = []string{
	"config",
	"win_config",
	"linux_config",
	"core_config",
	"win_core_config",
	"linux_core_config",
}

type ConsolePluginsFile struct {
	Url            string
	Destination    *string
	CollectionPath *string
}

type ConsolePlugin struct {
	Type  string
	Files []ConsolePluginsFile
}

type ConsoleLanguage struct {
	Tag  uint
	Name string
}

type ConsoleConfig struct {
	Name  string
	Value string
	Level string
}

type ConsoleFileType struct {
	FileType string
	Action   string
}

type Console struct {
	Slug                 string
	CoreLocation         string
	Name                 string
	SingleFile           bool
	IsEmbedded           bool
	LanguageVariableName *string
	Plugins              []ConsolePlugin
	FileTypes            []ConsoleFileType
	Configs              []ConsoleConfig
	Languages            []ConsoleLanguage
}

func PlainDatabaseToConsole(slug string, json interface{}) (console Console, err error) {
	var (
		entityObject map[string]interface{}
		ok           bool
	)
	if entityObject, ok = json.(map[string]interface{}); !ok {
		err = errors.New("the console JSON is not an object")
		return
	}
	if console, err = ConsoleFromJSON(slug, entityObject); err != nil {
		return
	}
	if consoleFileTypesObject, ok := entityObject["file_types"].(map[string]interface{}); ok {
		if err = PlainConsoleFileTypesToObject(&console, consoleFileTypesObject); err != nil {
			return
		}
	} else {
		err = errors.New("the console JSON not contains file types")
		return
	}
	if err = PlainConsoleConfigToObject(&console, entityObject); err != nil {
		return
	}
	if consoleLanguageObjectInterface, ok := entityObject["language"]; ok {
		if consoleLanguageObject, ok := consoleLanguageObjectInterface.(map[string]interface{}); ok {
			if err = PlainConsoleLanguageToObject(&console, consoleLanguageObject); err != nil {
				return
			}
		} else {
			err = errors.New("cannot parse language")
			return
		}
	}
	if consolePluginsObject, ok := entityObject["plugins"].(map[string]interface{}); ok {
		if err = PlainConsolePluginToObject(&console, consolePluginsObject); err != nil {
			return
		}
	}
	return
}

func PlainConsolePluginToObject(console *Console, consolePluginsObject map[string]interface{}) (err error) {
	for pluginKey, pluginValue := range consolePluginsObject {
		var consolePlugin ConsolePlugin
		consolePlugin, err = ConsolePluginFromJSON(pluginKey)
		console.Plugins = append(console.Plugins, consolePlugin)
		consolePluginObject := pluginValue.(map[string]interface{})
		if len(consolePluginObject) > 0 {
			consolePluginCollectionPath := consolePluginObject["collection_path"]
			consolePluginDestination := consolePluginObject["destination"]
			consolePluginFilesArray := consolePluginObject["files"].([]interface{})
			for fileIndex := 0; fileIndex < len(consolePluginFilesArray); fileIndex++ {
				var consolePluginCollectionPathValue interface{}
				if consolePluginCollectionPathObject, ok := consolePluginCollectionPath.([]interface{}); ok {
					consolePluginCollectionPathValue = consolePluginCollectionPathObject[fileIndex]
				} else {
					consolePluginCollectionPathValue = consolePluginCollectionPath
				}
				var consolePluginDestinationValue interface{}
				if consolePluginDestinationObject, ok := consolePluginDestination.([]interface{}); ok {
					consolePluginDestinationValue = consolePluginDestinationObject[fileIndex]
				} else {
					consolePluginDestinationValue = consolePluginDestination
				}
				var consolePluginsFile ConsolePluginsFile
				if consolePluginsFile, err = ConsolePluginsFileFromJSON(
					consolePluginCollectionPathValue,
					consolePluginDestinationValue,
					consolePluginFilesArray[fileIndex].(string)); err != nil {
					return
				}
				consolePlugin.Files = append(consolePlugin.Files, consolePluginsFile)
			}
		}
	}
	return
}

func PlainConsoleLanguageToObject(console *Console, consoleLanguageObject map[string]interface{}) (err error) {
	consoleLanguageMappingObject, _ := consoleLanguageObject["mapping"].(map[string]interface{})
	for languageIDKey, languageIDValue := range consoleLanguageMappingObject {
		for _, languageEntry := range languageIDValue.([]interface{}) {
			var languageID uint64
			if languageID, err = strconv.ParseUint(languageIDKey, 10, 32); err != nil {
				return
			}
			var consoleLanguage ConsoleLanguage
			if consoleLanguage, err = ConsoleLanguageFromJSON(uint(languageID), languageEntry.(string)); err != nil {
				return
			}
			console.Languages = append(console.Languages, consoleLanguage)
		}
	}
	return
}

func PlainConsoleConfigToObject(console *Console, entityObject map[string]interface{}) (err error) {
	for levelKey, levelValue := range entityObject {
		if ConsoleConfigIsLevel(levelKey) {
			consoleLevelObject := levelValue.(map[string]interface{})
			for consoleConfigName, consoleConfigValue := range consoleLevelObject {
				var consoleConfig ConsoleConfig
				if consoleConfig, err = ConsoleConfigFromJSON(levelKey, consoleConfigName, consoleConfigValue.(string)); err != nil {
					return
				}
				console.Configs = append(console.Configs, consoleConfig)
			}
		}
	}
	return
}

func PlainConsoleFileTypesToObject(console *Console, consoleFileTypesObject map[string]interface{}) (err error) {
	for actionKey, actionValue := range consoleFileTypesObject {
		for _, fileType := range actionValue.([]interface{}) {
			var consoleFileType ConsoleFileType
			if consoleFileType, err = ConsoleFileTypeFromJSON(actionKey, fileType.(string)); err != nil {
				return err
			}
			console.FileTypes = append(console.FileTypes, consoleFileType)
		}
	}
	return
}

func ConsoleFromJSON(slug string, json map[string]interface{}) (instance Console, err error) {
	var languageVariableName *string = nil
	if languageObject, ok := json["language"]; ok {
		if languageVariableNameObject, ok := languageObject.(map[string]interface{})["variable_name"]; ok {
			languageVariableNameVariable := languageVariableNameObject.(string)
			languageVariableName = &languageVariableNameVariable
		}
	}
	var (
		coreLocation string
		name         string
		singleFile   bool = true
		isEmbedded   bool = false
	)

	if value, ok := json["core_location"]; ok {
		coreLocation = value.(string)
	} else {
		err = errors.New("cannot parse core_location")
		return
	}
	if value, ok := json["name"]; ok {
		name = value.(string)
	} else {
		err = errors.New("cannot parse name")
		return
	}
	if value, ok := json["single_file"]; ok {
		singleFile = value.(bool)
	}
	if value, ok := json["is_embedded"]; ok {
		isEmbedded = value.(bool)
	}

	instance = Console{
		slug,
		coreLocation,
		name,
		singleFile,
		isEmbedded,
		languageVariableName,
		[]ConsolePlugin{},
		[]ConsoleFileType{},
		[]ConsoleConfig{},
		[]ConsoleLanguage{},
	}
	return
}

func ConsoleFileTypeFromJSON(actionString string, fileType string) (instance ConsoleFileType, err error) {
	instance = ConsoleFileType{
		fileType,
		actionString,
	}
	return
}

func ConsoleConfigIsLevel(level string) bool {
	for _, value := range consoleConfigLevels {
		if value == level {
			return true
		}
	}
	return false
}

func ConsoleConfigFromJSON(levelString string, name string, value string) (instance ConsoleConfig, err error) {
	instance = ConsoleConfig{
		name,
		value,
		levelString,
	}
	return
}

func ConsoleLanguageFromJSON(languageID uint, name string) (instance ConsoleLanguage, err error) {
	instance = ConsoleLanguage{
		languageID,
		name,
	}
	return
}

func ConsolePluginFromJSON(typeString string) (instance ConsolePlugin, err error) {
	instance = ConsolePlugin{
		typeString,
		[]ConsolePluginsFile{},
	}
	return
}

func ConsolePluginsFileFromJSON(jsonCollectionPath interface{}, jsonDestination interface{}, jsonFile string) (instance ConsolePluginsFile, err error) {
	var destination *string = nil
	if destinationObject, ok := jsonDestination.(string); ok {
		destination = &destinationObject
	}
	var collectionPath *string = nil
	if collectionPathObject, ok := jsonCollectionPath.(string); ok {
		collectionPath = &collectionPathObject
	}
	instance = ConsolePluginsFile{
		jsonFile,
		destination,
		collectionPath,
	}
	return
}
