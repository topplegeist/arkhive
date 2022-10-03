package importer

import (
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
	if console, err = ConsoleFromJSON(slug, json); err != nil {
		return
	}
	consoleObject := json.(map[string]interface{})
	consoleFileTypesObject, _ := consoleObject["file_types"].(map[string]interface{})
	for actionKey, actionValue := range consoleFileTypesObject {
		for _, fileType := range actionValue.([]interface{}) {
			var consoleFileType ConsoleFileType
			if consoleFileType, err = ConsoleFileTypeFromJSON(actionKey, fileType.(string)); err != nil {
				return
			}
			console.FileTypes = append(console.FileTypes, consoleFileType)
		}
	}
	for levelKey, levelValue := range consoleObject {
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
	if consoleLanguageObject, ok := consoleObject["language"].(map[string]interface{}); ok {
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
	}
	if consolePluginsObject, ok := consoleObject["plugins"].(map[string]interface{}); ok {
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
	}
	return
}

func ConsoleFromJSON(slug string, json interface{}) (instance Console, err error) {
	var languageVariableName *string = nil
	if languageObject, ok := json.(map[string]interface{})["language"]; ok {
		if languageVariableNameObject, ok := languageObject.(map[string]interface{})["variable_name"]; ok {
			languageVariableNameVariable := languageVariableNameObject.(string)
			languageVariableName = &languageVariableNameVariable
		}
	}
	instance = Console{
		slug,
		json.(map[string]interface{})["core_location"].(string),
		json.(map[string]interface{})["name"].(string),
		json.(map[string]interface{})["single_file"].(bool),
		json.(map[string]interface{})["is_embedded"].(bool),
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
