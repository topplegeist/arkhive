package engines

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"strconv"

	"arkhive.dev/launcher/common"
	"arkhive.dev/launcher/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DatabaseEngine struct {
	database *gorm.DB

	// Event emitters
	InitializationEndEventEmitter *common.EventEmitter
	DecryptedEventEmitter         *common.EventEmitter
}

func NewDatabaseEngine() (instance *DatabaseEngine, err error) {
	instance = new(DatabaseEngine)
	instance.InitializationEndEventEmitter = new(common.EventEmitter)
	instance.DecryptedEventEmitter = new(common.EventEmitter)

	go func() {
		if ok := instance.connectToDatabase(); !ok {
			log.Fatal("Cannot open database")
			return
		}
		if err = instance.applyMigrations(); err != nil {
			log.Fatal(err)
			return
		}
		storedDBHashString := instance.getStoredDBHash()
		var storedDBHash []byte
		if storedDBHashString != "" {
			if storedDBHash, err = base64.URLEncoding.DecodeString(storedDBHashString); err != nil {
				log.Fatal("Cannot decode the stored database hash")
				return
			}
		} else {
			log.Debug("Cannot get the stored database hash")
		}

		const cryptedDbFile = "db.bee"
		const plainDbFile = "db.json"
		const keyFilePath = "private_key.bee"
		_, existenceFlag := os.Stat(cryptedDbFile)
		cryptedDbFileExists := !os.IsNotExist(existenceFlag)
		_, existenceFlag = os.Stat(plainDbFile)
		plainDbFileExists := !os.IsNotExist(existenceFlag)
		_, existenceFlag = os.Stat(keyFilePath)
		keyFileExists := !os.IsNotExist(existenceFlag)
		hashHasBeenCalculated := len(storedDBHash) > 0

		canDecrypt := cryptedDbFileExists && keyFileExists
		plainAlreadyLoaded := plainDbFileExists && hashHasBeenCalculated
		if canDecrypt {
			var encryptedDBHash []byte
			if hashHasBeenCalculated {
				var encryptedDBData []byte
				if encryptedDBData, err = os.ReadFile(cryptedDbFile); err != nil {
					log.Fatal(err)
					return
				}
				hashEncoder := sha1.New()
				hashEncoder.Write(encryptedDBData)
				encryptedDBHash = hashEncoder.Sum(nil)
			}

			if !hashHasBeenCalculated || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
				var privateKey []byte
				if privateKey, err = os.ReadFile(keyFilePath); err != nil {
					log.Fatal(err)
					return
				}
				if privateKey, err = base64.URLEncoding.DecodeString(string(privateKey)); err != nil {
					log.Fatal("Cannot decode the stored encrypted database hash")
					return
				}
				var encryptedDBData []byte
				if encryptedDBData, err = os.ReadFile(cryptedDbFile); err != nil {
					log.Fatal(err)
					return
				}
				if encryptedDBData, err = base64.URLEncoding.DecodeString(string(encryptedDBData)); err != nil {
					log.Fatal("Cannot decode the stored encrypted database hash")
					return
				}
				if _, err = decode(encryptedDBData, privateKey); err != nil {
					log.Fatal("Cannot decode the encrypted database")
					return
				}
			}
		} else if !plainAlreadyLoaded {
			var dbData []byte
			if dbData, err = os.ReadFile(plainDbFile); err != nil {
				log.Fatal(err)
				return
			}

			hashEncoder := sha1.New()
			hashEncoder.Write(dbData)
			plainDBHash := hashEncoder.Sum(nil)

			if !hashHasBeenCalculated || !reflect.DeepEqual(storedDBHash, plainDBHash) {
				decoder := json.NewDecoder(bytes.NewReader(dbData))
				decoder.UseNumber()
				var db map[string]interface{}
				if err = decoder.Decode(&db); err != nil {
					log.Fatal(err)
					return
				}
				if err = instance.storeDecryptedDB(db); err != nil {
					log.Fatal(err)
					return
				}
			}

			storingDBHash := base64.URLEncoding.EncodeToString(plainDBHash)
			instance.setStoredDBHash(storingDBHash)
		}
		instance.InitializationEndEventEmitter.Emit(true)
	}()
	instance.InitializationEndEventEmitter.Subscribe(instance.databaseDecryptFutureFinished)
	return
}

func (databaseEngine *DatabaseEngine) connectToDatabase() bool {
	const fileName string = "data.sqllite3"
	var err error
	databaseEngine.database, err = gorm.Open(sqlite.Open(fileName), &gorm.Config{})
	return err == nil
}

func (databaseEngine DatabaseEngine) applyMigrations() (err error) {
	err = databaseEngine.database.AutoMigrate(&models.User{},
		&models.Chat{}, &models.Tool{}, &models.Console{}, &models.Game{},
		&models.ToolFilesType{}, &models.ConsoleFileType{}, &models.ConsoleLanguage{},
		&models.ConsolePlugin{}, &models.ConsolePluginsFile{},
		&models.ConsoleConfig{}, &models.GameDisk{}, &models.GameAdditionalFile{},
		&models.GameConfig{}, &models.UserVariable{})
	return
}

func (databaseEngine DatabaseEngine) getStoredDBHash() string {
	var userVariable models.UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "dbHash"); result.Error != nil || !userVariable.Value.Valid {
		return ""
	}
	return userVariable.Value.String
}

func (databaseEngine DatabaseEngine) setStoredDBHash(dbHash string) {
	userVariable := models.UserVariable{
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

func (databaseEngine DatabaseEngine) storeDecryptedDB(database map[string]interface{}) (err error) {
	if err = databaseEngine.storeDecryptedConsoles(database["consoles"].(map[string]interface{})); err != nil {
		return
	}
	if err = databaseEngine.storeDecryptedGames(database["games"].(map[string]interface{})); err != nil {
		return
	}
	if err = databaseEngine.storeDecryptedTools(database["win_tools"].(map[string]interface{})); err != nil {
		return
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedConsoles(consolesJson map[string]interface{}) (err error) {
	for consoleKey, consoleValue := range consolesJson {
		var console *models.Console
		if console, err = models.ConsoleFromJSON(consoleKey, consoleValue); err != nil {
			return
		}
		databaseEngine.database.Create(console)
		consoleObject := consoleValue.(map[string]interface{})
		consoleFileTypesObject, _ := consoleObject["file_types"].(map[string]interface{})
		for actionKey, actionValue := range consoleFileTypesObject {
			for _, fileType := range actionValue.([]interface{}) {
				var consoleFileType *models.ConsoleFileType
				if consoleFileType, err = models.ConsoleFileTypeFromJSON(actionKey, console, fileType); err != nil {
					return
				}
				databaseEngine.database.Create(consoleFileType)
			}
		}
		for levelKey, levelValue := range consoleObject {
			if models.ConsoleConfigIsLevel(levelKey) {
				consoleLevelObject := levelValue.(map[string]interface{})
				for consoleConfigName, consoleConfigValue := range consoleLevelObject {
					var consoleConfig *models.ConsoleConfig
					if consoleConfig, err = models.ConsoleConfigFromJSON(console, levelKey, consoleConfigName, consoleConfigValue); err != nil {
						return
					}
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
					var consoleLanguage *models.ConsoleLanguage
					if consoleLanguage, err = models.ConsoleLanguageFromJSON(console, uint(languageID), languageEntry); err != nil {
						return
					}
					databaseEngine.database.Create(consoleLanguage)
				}
			}
		}
		if consolePluginsObject, ok := consoleObject["plugins"].(map[string]interface{}); ok {
			for pluginKey, pluginValue := range consolePluginsObject {
				var consolePlugin *models.ConsolePlugin
				consolePlugin, err = models.ConsolePluginFromJSON(pluginKey, console)
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
						var consolePluginsFile *models.ConsolePluginsFile
						if consolePluginsFile, err = models.ConsolePluginsFileFromJSON(
							consolePlugin, consolePluginCollectionPathValue,
							consolePluginDestinationValue,
							consolePluginFilesArray[fileIndex]); err != nil {
							return
						}
						databaseEngine.database.Create(consolePluginsFile)
					}
				}
			}
		}
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedGames(gamesJson map[string]interface{}) (err error) {
	for gameKey, gameValue := range gamesJson {
		var console models.Console
		gameObject := gameValue.(map[string]interface{})
		if result := databaseEngine.database.First(&console, "slug = ?", gameObject["console_slug"].(string)); result.Error != nil {
			err = result.Error
			return
		}
		var game *models.Game
		if game, err = models.GameFromJSON(gameKey, &console, gameValue); err != nil {
			return
		}
		databaseEngine.database.Create(game)
		collectionPath := gameObject["collection_path"]
		var gameDisk *models.GameDisk
		if gameUrls, ok := gameObject["url"].([]interface{}); ok {
			for diskNumber := 0; diskNumber < len(gameUrls); diskNumber++ {
				gameDiskImage := gameObject["disk_image"].([]interface{})[diskNumber]
				if gameDisk, err = models.GameDiskFromJSON(game, uint(diskNumber), gameUrls[diskNumber], gameDiskImage, collectionPath); err != nil {
					return
				}
				databaseEngine.database.Create(gameDisk)
			}
		} else {
			if gameDisk, err = models.GameDiskFromJSON(game, 0, gameObject["url"], nil, collectionPath); err != nil {
				return
			}
			databaseEngine.database.Create(gameDisk)
		}
		if gameConfigObject, ok := gameObject["config"].(map[string]interface{}); ok {
			for configKey, configValue := range gameConfigObject {
				var gameConfig *models.GameConfig
				if gameConfig, err = models.GameConfigFromJSON(game, configKey, configValue); err != nil {
					return
				}
				databaseEngine.database.Create(gameConfig)
			}
		}
		if gameAdditionlFilesObject, ok := gameObject["additional_files"].([]interface{}); ok {
			for _, gameAdditionlFileObject := range gameAdditionlFilesObject {
				var gameAdditionalFile *models.GameAdditionalFile
				if gameAdditionalFile, err = models.GameAdditionalFileFromJSON(game, gameAdditionlFileObject); err != nil {
					return
				}
				databaseEngine.database.Create(gameAdditionalFile)
			}
		}
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedTools(toolsJson map[string]interface{}) (err error) {
	for toolKey, toolValue := range toolsJson {
		var tool *models.Tool
		if tool, err = models.ToolFromJSON(toolKey, toolsJson[toolKey]); err != nil {
			return
		}
		databaseEngine.database.Create(tool)
		if toolFileTypesObject, ok := toolValue.(map[string]interface{})["file_types"].([]interface{}); ok {
			for _, toolFileTypeObject := range toolFileTypesObject {
				var toolFileType *models.ToolFilesType
				if toolFileType, err = models.ToolFilesTypeFromJSON(tool, toolFileTypeObject); err != nil {
					return
				}
				databaseEngine.database.Create(toolFileType)
			}
		}
	}
	return
}

func (databaseEngine DatabaseEngine) databaseDecryptFutureFinished(hasBeenInitializated bool) {
	if hasBeenInitializated {
		databaseEngine.DecryptedEventEmitter.Emit(true)
	}
}
