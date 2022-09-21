package engine

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"sync"

	"arkhive.dev/launcher/internal/console"
	"arkhive.dev/launcher/internal/database/delegate"
	"arkhive.dev/launcher/internal/database/importer"
	"arkhive.dev/launcher/internal/entity"
	"arkhive.dev/launcher/internal/game"
	"arkhive.dev/launcher/internal/tool"
	"github.com/sirupsen/logrus"
)

type Locale int

const (
	ENGLISH Locale = iota
	FRENCH
	SPANISH
	GERMAN
	ITALIAN
)

type Database struct {
	basePath  string
	delegate  delegate.DatabaseDelegate
	importers []importer.Importer
}

func NewDatabase(basePath string, delegate delegate.DatabaseDelegate, importers []importer.Importer) (instance *Database) {
	instance = &Database{
		basePath:  basePath,
		delegate:  delegate,
		importers: importers,
	}
	return
}

func (databaseEngine *Database) Initialize(waitGroup *sync.WaitGroup) {
	var err error
	// Create or update the database if needed
	logrus.Info("Connecting to database")
	if ok := databaseEngine.connectToDatabase(); !ok {
		panic("cannot open database")
	}
	logrus.Info("Applying database migrations")
	if err = databaseEngine.applyMigrations(); err != nil {
		panic(err)
	}

	// Check whether the database hash has been already saved on the database
	var storedDBHash []byte
	if storedDBHash, err = databaseEngine.getStoredDBHash(); err != nil {
		logrus.Error("Cannot decode the stored database hash")
		panic(err)
	}
	hashHasBeenStored := len(storedDBHash) > 0
	if !hashHasBeenStored {
		logrus.Info("Cannot get the stored database hash")
	}

	// Import the database from the higher priority importer to the lower
	var (
		databaseData    []byte
		encryptedDBHash []byte
	)
	for _, importer := range databaseEngine.importers {
		if importer.CanImport() {
			if databaseData, encryptedDBHash, err = importer.Import(); err != nil {
				panic("error importing the database")
			}
			break
		}
	}

	// Parse the database data read, if any
	if databaseData != nil {
		logrus.Info("Storing the database")
		if err = databaseEngine.storeDecryptedDatabase(databaseData); err != nil {
			panic(err)
		}
		storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
		databaseEngine.setStoredDBHash(storingDBHash)
	}

	// End the routine
	waitGroup.Done()
}

func (databaseEngine *Database) Deinitialize() {
	databaseEngine.delegate.Close()
}

func (databaseEngine *Database) connectToDatabase() bool {
	var err error
	err = databaseEngine.delegate.Open(databaseEngine.basePath)
	return err == nil
}

func (databaseEngine Database) applyMigrations() (err error) {
	if err = databaseEngine.database.AutoMigrate(&entity.User{},
		&entity.Chat{}, &tool.Tool{}, &console.Console{}, &game.Game{},
		&tool.ToolFilesType{}, &console.ConsoleFileType{}, &console.ConsoleLanguage{},
		&console.ConsolePlugin{}, &console.ConsolePluginsFile{},
		&console.ConsoleConfig{}, &game.GameDisk{}, &game.GameAdditionalFile{},
		&game.GameConfig{}, &entity.UserVariable{}); err != nil {
		return err
	}
	if err = databaseEngine.delegate.Migrate(); err != nil {
		return err
	}
	return
}

func (databaseEngine Database) storeDecryptedDatabase(dbData []byte) (err error) {
	decoder := json.NewDecoder(bytes.NewReader(dbData))
	decoder.UseNumber()
	var database map[string]interface{}
	if err = decoder.Decode(&database); err != nil {
		logrus.Errorf("%+v", err)
		return
	}

	if entities, ok := database["consoles"]; ok {
		if err = databaseEngine.storeDecryptedConsoles(entities.(map[string]interface{})); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
	}
	if entities, ok := database["games"]; ok {
		if err = databaseEngine.storeDecryptedGames(entities.(map[string]interface{})); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
	}
	if entities, ok := database["win_tools"]; ok {
		if err = databaseEngine.storeDecryptedTools(entities.(map[string]interface{})); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
	}
	return
}

func (databaseEngine Database) storeDecryptedConsoles(consolesJson map[string]interface{}) (err error) {
	for consoleKey, consoleValue := range consolesJson {
		var consoleInstance *console.Console
		if consoleInstance, err = console.ConsoleFromJSON(consoleKey, consoleValue); err != nil {
			return
		}
		logrus.Infof("Storing %s", consoleInstance.Slug)
		databaseEngine.database.Create(consoleInstance)
		consoleObject := consoleValue.(map[string]interface{})
		consoleFileTypesObject, _ := consoleObject["file_types"].(map[string]interface{})
		for actionKey, actionValue := range consoleFileTypesObject {
			for _, fileType := range actionValue.([]interface{}) {
				var consoleFileType *console.ConsoleFileType
				if consoleFileType, err = console.ConsoleFileTypeFromJSON(actionKey, consoleInstance, fileType.(string)); err != nil {
					return
				}
				logrus.Infof("Storing %s %s file type", consoleInstance.Slug, consoleFileType.FileType)
				databaseEngine.database.Create(consoleFileType)
			}
		}
		for levelKey, levelValue := range consoleObject {
			if console.ConsoleConfigIsLevel(levelKey) {
				consoleLevelObject := levelValue.(map[string]interface{})
				for consoleConfigName, consoleConfigValue := range consoleLevelObject {
					var consoleConfig *console.ConsoleConfig
					if consoleConfig, err = console.ConsoleConfigFromJSON(consoleInstance, levelKey, consoleConfigName, consoleConfigValue.(string)); err != nil {
						return
					}
					logrus.Infof("Storing %s %s configuration", consoleInstance.Slug, consoleConfig.Name)
					databaseEngine.database.Create(consoleConfig)
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
					var consoleLanguage *console.ConsoleLanguage
					if consoleLanguage, err = console.ConsoleLanguageFromJSON(consoleInstance, uint(languageID), languageEntry.(string)); err != nil {
						return
					}
					logrus.Infof("Storing %s %s language", consoleInstance.Slug, consoleLanguage.Name)
					databaseEngine.database.Create(consoleLanguage)
				}
			}
		}
		if consolePluginsObject, ok := consoleObject["plugins"].(map[string]interface{}); ok {
			for pluginKey, pluginValue := range consolePluginsObject {
				var consolePlugin *console.ConsolePlugin
				consolePlugin, err = console.ConsolePluginFromJSON(pluginKey, consoleInstance)
				logrus.Infof("Storing %s %s plugin", consoleInstance.Slug, consolePlugin.Type)
				if databaseEngine.database.Create(consolePlugin); err != nil {
					return
				}
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
						var consolePluginsFile *console.ConsolePluginsFile
						if consolePluginsFile, err = console.ConsolePluginsFileFromJSON(
							consolePlugin, consolePluginCollectionPathValue,
							consolePluginDestinationValue,
							consolePluginFilesArray[fileIndex].(string)); err != nil {
							return
						}
						logrus.Infof("Storing %s %s plugin %s file", consoleInstance.Slug, consolePlugin.Type, consolePluginsFile.Url)
						databaseEngine.database.Create(consolePluginsFile)
					}
				}
			}
		}
	}
	return
}

func (databaseEngine Database) storeDecryptedGames(gamesJson map[string]interface{}) (err error) {
	for gameKey, gameValue := range gamesJson {
		var consoleInstance console.Console
		gameObject := gameValue.(map[string]interface{})
		if result := databaseEngine.database.First(&consoleInstance, "slug = ?", gameObject["console_slug"].(string)); result.Error != nil {
			err = result.Error
			return
		}
		var gameInstance *game.Game
		if gameInstance, err = game.GameFromJSON(gameKey, &consoleInstance, gameValue); err != nil {
			return
		}
		logrus.Infof("Storing %s game", gameInstance.Slug)
		databaseEngine.database.Create(gameInstance)

		collectionPath := gameObject["collection_path"]
		var gameDisks = []*game.GameDisk{}
		if gameUrls, ok := gameObject["url"].([]interface{}); ok {
			for diskNumber := 0; diskNumber < len(gameUrls); diskNumber++ {
				var gameDisk *game.GameDisk
				gameDiskImage := gameObject["disk_image"].([]interface{})[diskNumber]
				if gameDisk, err = game.GameDiskFromJSON(gameInstance, uint(diskNumber), gameUrls[diskNumber].(string), gameDiskImage, collectionPath); err != nil {
					return
				}
				gameDisks = append(gameDisks, gameDisk)
			}
		} else {
			var gameDisk *game.GameDisk
			if gameDisk, err = game.GameDiskFromJSON(gameInstance, 0, gameObject["url"].(string), nil, collectionPath); err != nil {
				return
			}
			gameDisks = append(gameDisks, gameDisk)
		}
		for _, gameDisk := range gameDisks {
			logrus.Infof("Storing %s game disk %d", gameInstance.Slug, gameDisk.DiskNumber)
			databaseEngine.database.Create(gameDisk)
		}

		if gameConfigObject, ok := gameObject["config"].(map[string]interface{}); ok {
			for configKey, configValue := range gameConfigObject {
				var gameConfig *game.GameConfig
				if gameConfig, err = game.GameConfigFromJSON(gameInstance, configKey, configValue); err != nil {
					return
				}
				logrus.Infof("Storing %s %s configuration", gameInstance.Slug, gameConfig.Name)
				databaseEngine.database.Create(gameConfig)
			}
		}
		if gameAdditionlFilesObject, ok := gameObject["additional_files"].([]interface{}); ok {
			for _, gameAdditionlFileObject := range gameAdditionlFilesObject {
				var gameAdditionalFile *game.GameAdditionalFile
				if gameAdditionalFile, err = game.GameAdditionalFileFromJSON(gameInstance, gameAdditionlFileObject); err != nil {
					return
				}
				logrus.Infof("Storing %s %s additional File", gameInstance.Slug, gameAdditionalFile.Name)
				databaseEngine.database.Create(gameAdditionalFile)
			}
		}
	}
	return
}

func (databaseEngine Database) storeDecryptedTools(toolsJson map[string]interface{}) (err error) {
	for toolKey, toolValue := range toolsJson {
		var toolInstance *tool.Tool
		if toolInstance, err = tool.ToolFromJSON(toolKey, toolsJson[toolKey]); err != nil {
			return
		}
		logrus.Infof("Storing %s tool", toolInstance.Slug)
		databaseEngine.database.Create(toolInstance)

		if toolFileTypesObject, ok := toolValue.(map[string]interface{})["file_types"].([]interface{}); ok {
			for _, toolFileTypeObject := range toolFileTypesObject {
				var toolFileType *tool.ToolFilesType
				if toolFileType, err = tool.ToolFilesTypeFromJSON(toolInstance, toolFileTypeObject); err != nil {
					return
				}
				logrus.Infof("Storing %s %s tool file", toolInstance.Slug, toolFileType.Tool.Url)
				databaseEngine.database.Create(toolFileType)
			}
		}
	}
	return
}

func (databaseEngine Database) getStoredDBHash() (storedDBHash []byte, err error) {
	var userVariable entity.UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "dbHash"); result.Error != nil || !userVariable.Value.Valid {
		storedDBHash = []byte{}
		return
	}
	storedDBHash, err = base64.URLEncoding.DecodeString(userVariable.Value.String)
	return
}

func (databaseEngine Database) setStoredDBHash(dbHash string) {
	userVariable := entity.UserVariable{
		Name: "dbHash",
		Value: sql.NullString{
			String: dbHash,
			Valid:  true,
		},
	}
	databaseEngine.database.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&userVariable)
}
