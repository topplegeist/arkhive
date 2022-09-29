package engine

import (
	"bytes"
	"database/sql"
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

func (d *Database) Initialize(waitGroup *sync.WaitGroup) {
	var err error
	// Create or update the database if needed
	logrus.Info("Connecting to database")
	if ok := d.connectToDatabase(); !ok {
		panic("cannot open database")
	}
	logrus.Info("Applying database migrations")
	if err = d.applyMigrations(); err != nil {
		panic(err)
	}

	// Check whether the database hash has been already saved on the database
	var storedDBHash []byte
	if storedDBHash, err = d.getStoredDBHash(); err != nil {
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
	for _, importer := range d.importers {
		if importer.CanImport() && importer.IsUpdated() {
			if databaseData, encryptedDBHash, err = importer.Import(); err != nil {
				panic("error importing the database")
			}
			break
		}
	}

	// Parse the database data read, if any
	if databaseData != nil {
		logrus.Info("Storing the database")
		if err = d.storeDecryptedDatabase(databaseData); err != nil {
			panic(err)
		}
		storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
		d.setStoredDBHash(storingDBHash)
	}

	// End the routine
	waitGroup.Done()
}

func (d *Database) Deinitialize() {
	d.delegate.Close()
}

func (d *Database) connectToDatabase() bool {
	return d.delegate.Open(d.basePath) == nil
}

func (d Database) applyMigrations() (err error) {
	if err = d.delegate.Migrate(); err != nil {
		return err
	}
	return
}

func (d Database) storeDecryptedDatabase(dbData []byte) (err error) {
	decoder := json.NewDecoder(bytes.NewReader(dbData))
	decoder.UseNumber()
	var database map[string]interface{}
	if err = decoder.Decode(&database); err != nil {
		logrus.Errorf("%+v", err)
		return
	}

	if entities, ok := database["consoles"]; ok {
		if err = d.storeDecryptedConsoles(entities.(map[string]interface{})); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
	}
	if entities, ok := database["games"]; ok {
		if err = d.storeDecryptedGames(entities.(map[string]interface{})); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
	}
	if entities, ok := database["win_tools"]; ok {
		if err = d.storeDecryptedTools(entities.(map[string]interface{})); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
	}
	return
}

func (d Database) extract

func (d Database) storeDecryptedConsoles(consolesJson map[string]interface{}) (err error) {
	for consoleKey, consoleValue := range consolesJson {
		var consoleInstance *console.Console
		if consoleInstance, err = console.ConsoleFromJSON(consoleKey, consoleValue); err != nil {
			return
		}
		logrus.Infof("Storing %s", consoleInstance.Slug)
		if err = d.delegate.Create(consoleInstance); err != nil {
			return
		}
		consoleObject := consoleValue.(map[string]interface{})
		consoleFileTypesObject, _ := consoleObject["file_types"].(map[string]interface{})
		for actionKey, actionValue := range consoleFileTypesObject {
			for _, fileType := range actionValue.([]interface{}) {
				var consoleFileType *console.ConsoleFileType
				if consoleFileType, err = console.ConsoleFileTypeFromJSON(actionKey, consoleInstance, fileType.(string)); err != nil {
					return
				}
				logrus.Infof("Storing %s %s file type", consoleInstance.Slug, consoleFileType.FileType)
				if err = d.delegate.Create(consoleFileType); err != nil {
					return
				}
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
					if err = d.delegate.Create(consoleConfig); err != nil {
						return
					}
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
					if err = d.delegate.Create(consoleLanguage); err != nil {
						return
					}
				}
			}
		}
		if consolePluginsObject, ok := consoleObject["plugins"].(map[string]interface{}); ok {
			for pluginKey, pluginValue := range consolePluginsObject {
				var consolePlugin *console.ConsolePlugin
				consolePlugin, err = console.ConsolePluginFromJSON(pluginKey, consoleInstance)
				logrus.Infof("Storing %s %s plugin", consoleInstance.Slug, consolePlugin.Type)
				if d.delegate.Create(consolePlugin); err != nil {
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
						if err = d.delegate.Create(consolePluginsFile); err != nil {
							return
						}
					}
				}
			}
		}
	}
	return
}

func (d Database) storeDecryptedGames(gamesJson map[string]interface{}) (err error) {
	for gameKey, gameValue := range gamesJson {
		var consoleInstance console.Console
		gameObject := gameValue.(map[string]interface{})
		if err = d.delegate.First(&consoleInstance, "slug = ?", gameObject["console_slug"].(string)); err != nil {
			return
		}
		var gameInstance *game.Game
		if gameInstance, err = game.GameFromJSON(gameKey, &consoleInstance, gameValue); err != nil {
			return
		}
		logrus.Infof("Storing %s game", gameInstance.Slug)
		if err = d.delegate.Create(gameInstance); err != nil {
			return
		}

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
			if err = d.delegate.Create(gameDisk); err != nil {
				return
			}
		}

		if gameConfigObject, ok := gameObject["config"].(map[string]interface{}); ok {
			for configKey, configValue := range gameConfigObject {
				var gameConfig *game.GameConfig
				if gameConfig, err = game.GameConfigFromJSON(gameInstance, configKey, configValue); err != nil {
					return
				}
				logrus.Infof("Storing %s %s configuration", gameInstance.Slug, gameConfig.Name)
				if err = d.delegate.Create(gameConfig); err != nil {
					return
				}
			}
		}
		if gameAdditionlFilesObject, ok := gameObject["additional_files"].([]interface{}); ok {
			for _, gameAdditionlFileObject := range gameAdditionlFilesObject {
				var gameAdditionalFile *game.GameAdditionalFile
				if gameAdditionalFile, err = game.GameAdditionalFileFromJSON(gameInstance, gameAdditionlFileObject); err != nil {
					return
				}
				logrus.Infof("Storing %s %s additional File", gameInstance.Slug, gameAdditionalFile.Name)
				if err = d.delegate.Create(gameAdditionalFile); err != nil {
					return
				}
			}
		}
	}
	return
}

func (d Database) storeDecryptedTools(toolsJson map[string]interface{}) (err error) {
	for toolKey, toolValue := range toolsJson {
		var toolInstance *tool.Tool
		if toolInstance, err = tool.ToolFromJSON(toolKey, toolsJson[toolKey]); err != nil {
			return
		}
		logrus.Infof("Storing %s tool", toolInstance.Slug)
		if err = d.delegate.Create(toolInstance); err != nil {
			return
		}

		if toolFileTypesObject, ok := toolValue.(map[string]interface{})["file_types"].([]interface{}); ok {
			for _, toolFileTypeObject := range toolFileTypesObject {
				var toolFileType *tool.ToolFilesType
				if toolFileType, err = tool.ToolFilesTypeFromJSON(toolInstance, toolFileTypeObject); err != nil {
					return
				}
				logrus.Infof("Storing %s %s tool file", toolInstance.Slug, toolFileType.Tool.Url)
				if err = d.delegate.Create(toolFileType); err != nil {
					return
				}
			}
		}
	}
	return
}

func (d Database) getStoredDBHash() (storedDBHash []byte, err error) {
	var userVariable entity.UserVariable
	if err = d.delegate.First(&userVariable, "name = ?", "dbHash"); err != nil || !userVariable.Value.Valid {
		storedDBHash = []byte{}
		return
	}
	storedDBHash, err = base64.URLEncoding.DecodeString(userVariable.Value.String)
	return
}

func (d Database) setStoredDBHash(dbHash string) (err error) {
	userVariable := entity.UserVariable{
		Name: "dbHash",
		Value: sql.NullString{
			String: dbHash,
			Valid:  true,
		},
	}
	if err = d.delegate.CreateOrUpdate(&userVariable); err != nil {
		return
	}
	return
}
