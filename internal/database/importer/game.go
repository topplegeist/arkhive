package importer

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

type GameAdditionalFile struct {
	Name string
	Data []byte
}

type GameConfig struct {
	Name  string
	Value string
}

type GameDisk struct {
	DiskNumber     uint
	Url            string
	Image          *string
	CollectionPath *string
}

type Game struct {
	Slug            string
	Name            string
	ConsoleSlug     string
	BackgroundColor string
	BackgroundImage *string
	Logo            *string
	Executable      *string
	Disks           []GameDisk
	Configs         []GameConfig
	AdditionalFiles []GameAdditionalFile
}

func PlainDatabaseToGame(slug string, json interface{}) (game Game, err error) {
	var (
		entityObject map[string]interface{}
		ok           bool
	)
	if entityObject, ok = json.(map[string]interface{}); !ok {
		err = errors.New("the game JSON is not an object")
		return
	}

	if game, err = GameFromJSON(slug, entityObject); err != nil {
		return
	}

	collectionPath := entityObject["collection_path"]
	if urls, ok := entityObject["url"].([]interface{}); ok {
		for diskNumber := 0; diskNumber < len(urls); diskNumber++ {
			var disk GameDisk
			diskImage := entityObject["disk_image"].([]interface{})[diskNumber]
			if disk, err = GameDiskFromJSON(uint(diskNumber), urls[diskNumber].(string), diskImage, collectionPath); err != nil {
				return
			}
			game.Disks = append(game.Disks, disk)
		}
	} else {
		var disk GameDisk
		if disk, err = GameDiskFromJSON(0, entityObject["url"].(string), nil, collectionPath); err != nil {
			return
		}
		game.Disks = append(game.Disks, disk)
	}

	if configObject, ok := entityObject["config"].(map[string]interface{}); ok {
		for configKey, configValue := range configObject {
			var config GameConfig
			if config, err = GameConfigFromJSON(configKey, configValue); err != nil {
				return
			}
			game.Configs = append(game.Configs, config)
		}
	}
	if additionalFilesObject, ok := entityObject["additional_files"].([]interface{}); ok {
		for _, additionalFileObject := range additionalFilesObject {
			var additionalFile GameAdditionalFile
			if additionalFile, err = GameAdditionalFileFromJSON(additionalFileObject); err != nil {
				return
			}
			game.AdditionalFiles = append(game.AdditionalFiles, additionalFile)
		}
	}
	return
}

func GameFromJSON(slug string, data map[string]interface{}) (instance Game, err error) {
	var backgroundImage *string
	if backgroundImageObject, ok := data["background_image"].(string); ok {
		backgroundImage = &backgroundImageObject
	}
	var logo *string
	if logoObject, ok := data["logo"].(string); ok {
		logo = &logoObject
	}
	var executable *string
	if executableObject, ok := data["executable"].(string); ok {
		executable = &executableObject
	}
	instance = Game{
		slug,
		data["name"].(string),
		data["console_slug"].(string),
		data["background_color"].(string),
		backgroundImage,
		logo,
		executable,
		[]GameDisk{},
		[]GameConfig{},
		[]GameAdditionalFile{},
	}
	return
}

func GameDiskFromJSON(diskNumber uint, jsonUrl string, jsonDiskImage interface{}, jsonCollectionPath interface{}) (instance GameDisk, err error) {
	var image *string
	if imageObject, ok := jsonDiskImage.(string); ok {
		image = &imageObject
	}
	var collectionPath *string
	if collectionPathObject, ok := jsonCollectionPath.(string); ok {
		collectionPath = &collectionPathObject
	}
	instance = GameDisk{
		diskNumber,
		jsonUrl,
		image,
		collectionPath,
	}
	return
}

func GameConfigFromJSON(name string, jsonValue interface{}) (instance GameConfig, err error) {
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

	instance = GameConfig{
		name,
		value,
	}
	return
}

func GameAdditionalFileFromJSON(json interface{}) (instance GameAdditionalFile, err error) {
	var data []byte
	if data, err = base64.URLEncoding.DecodeString(json.(map[string]interface{})["base64"].(string)); err != nil {
		return
	}
	instance = GameAdditionalFile{
		json.(map[string]interface{})["name"].(string),
		data,
	}
	return
}
