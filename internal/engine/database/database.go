package database

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strconv"

	"arkhive.dev/launcher/internal/engine/database/importer"
	"arkhive.dev/launcher/internal/entity"
	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/encryption"
	"arkhive.dev/launcher/pkg/eventemitter"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Locale int

const (
	ENGLISH Locale = iota
	FRENCH
	SPANISH
	GERMAN
	ITALIAN
)

type DatabaseEngine struct {
	database *gorm.DB

	// Event emitters
	BootedEventEmitter    *eventemitter.EventEmitter
	DecryptedEventEmitter *eventemitter.EventEmitter
}

func NewDatabaseEngine() (instance *DatabaseEngine) {
	instance = new(DatabaseEngine)
	instance.BootedEventEmitter = new(eventemitter.EventEmitter)
	instance.DecryptedEventEmitter = new(eventemitter.EventEmitter)
	return
}

func (databaseEngine *DatabaseEngine) Initialize() (err error) {
	log.Info("Connecting to database")
	if ok := databaseEngine.connectToDatabase(); !ok {
		panic("Cannot open database")
		return
	}
	log.Info("Applying database migrations")
	if err = databaseEngine.applyMigrations(); err != nil {
		panic(err)
		return
	}

	var storedDBHash []byte
	if storedDBHash, err = databaseEngine.getStoredDBHash(); err != nil {
		log.Error("Cannot decode the stored database hash")
		panic(err)
		return
	}
	hashHasBeenStored := len(storedDBHash) > 0
	if !hashHasBeenStored {
		log.Info("Cannot get the stored database hash")
	}

	_, existenceFlag := os.Stat(folder.CryptedDatabasePath)
	cryptedDbFileExists := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(folder.PlainDatabasePath)
	plainDbFileExists := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(folder.DatabaseKeyPath)
	keyFileExists := !os.IsNotExist(existenceFlag)

	canDecrypt := cryptedDbFileExists && keyFileExists
	var encryptedDBHash []byte

	if canDecrypt {
		log.Info("Loading database private key")
		var privateKey *rsa.PrivateKey
		if privateKey, err = databaseEngine.loadPrivateKey(); err != nil {
			panic(err)
		}
		var cryptedDatabaseReader io.Reader
		if cryptedDatabaseReader, err = os.Open(folder.CryptedDatabasePath); err != nil {
			log.Error("Cannot read the database key file")
			panic(err)
		}
		var (
			databaseData    []byte
			encryptedDBHash []byte
		)
		if databaseData, encryptedDBHash, err = importer.ImportCryptedDatabase(cryptedDatabaseReader, privateKey, storedDBHash); err != nil {
			panic(err)
		}

		if databaseData != nil {
			log.Info("Storing the database")
			if err = databaseEngine.storeDecryptedDatabase(databaseData); err != nil {
				panic(err)
			}
			storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
			databaseEngine.setStoredDBHash(storingDBHash)
		}
	} else if plainDbFileExists {
		log.Info("The encrypted database cannot be decrypted, proceeding with the plain JSON file")
		var plainDatabaseFileReader *os.File
		if plainDatabaseFileReader, err = os.Open(folder.PlainDatabasePath); err != nil {
			log.Error("Cannot read the plain database file")
			panic(err)
		}
		var (
			databaseKeyReader io.Reader
			databaseKeyWriter io.Writer
			undertowWriter    io.Writer
		)
		if keyFileExists {
			if databaseKeyReader, err = os.Open(folder.DatabaseKeyPath); err != nil {
				log.Error("Cannot read the database key file")
				panic(err)
			}
		} else {
			if databaseKeyWriter, err = os.Create(folder.DatabaseKeyPath); err != nil {
				log.Error("Cannot create the database key file")
				panic(err)
			}
			if undertowWriter, err = os.Create(folder.NewUndertowPath); err != nil {
				log.Error("Cannot create the new undertow file")
				panic(err)
			}
		}
		var cryptedDatabaseWriter io.Writer
		if cryptedDbFileExists {
			os.Remove(folder.CryptedDatabasePath)
		}
		if cryptedDatabaseWriter, err = os.Create(folder.CryptedDatabasePath); err != nil {
			log.Error("Cannot create the crypted database file")
			panic(err)
		}

		var databaseData []byte
		if databaseData, encryptedDBHash, err = importer.ImportPlainDatabase(
			plainDatabaseFileReader, databaseKeyReader, databaseKeyWriter, undertowWriter, cryptedDatabaseWriter, cryptedDbFileExists); err != nil {
			panic(err)
		}

		if !hashHasBeenStored || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
			if err = databaseEngine.storeDecryptedDatabase(databaseData); err != nil {
				panic(err)
			}
			storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
			databaseEngine.setStoredDBHash(storingDBHash)
		}
	} else if !hashHasBeenStored {
		panic("No database to be imported")
	}

	databaseEngine.DecryptedEventEmitter.Emit(true)
	databaseEngine.BootedEventEmitter.Emit(true)
	return
}

func (databaseEngine *DatabaseEngine) Deinitialize() {
	if database, err := databaseEngine.database.DB(); err == nil {
		database.Close()
	}
}

func (databaseEngine *DatabaseEngine) connectToDatabase() bool {
	var err error
	databaseEngine.database, err = gorm.Open(sqlite.Open(folder.DatabasePath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	return err == nil
}

func (databaseEngine DatabaseEngine) applyMigrations() (err error) {
	err = databaseEngine.database.AutoMigrate(&entity.User{},
		&entity.Chat{}, &entity.Tool{}, &entity.Console{}, &entity.Game{},
		&entity.ToolFilesType{}, &entity.ConsoleFileType{}, &entity.ConsoleLanguage{},
		&entity.ConsolePlugin{}, &entity.ConsolePluginsFile{},
		&entity.ConsoleConfig{}, &entity.GameDisk{}, &entity.GameAdditionalFile{},
		&entity.GameConfig{}, &entity.UserVariable{})
	return
}

func (databaseEngine *DatabaseEngine) loadPrivateKey() (privateKey *rsa.PrivateKey, err error) {
	var privateKeyBytes []byte
	if privateKeyBytes, err = os.ReadFile(folder.DatabaseKeyPath); err != nil {
		log.Error("Cannot read the secret key file")
		return
	}
	if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
		log.Error("Cannot import the private key")
		return
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedDatabase(dbData []byte) (err error) {
	decoder := json.NewDecoder(bytes.NewReader(dbData))
	decoder.UseNumber()
	var database map[string]interface{}
	if err = decoder.Decode(&database); err != nil {
		log.Error(err)
		return
	}

	if err = databaseEngine.storeDecryptedConsoles(database["consoles"].(map[string]interface{})); err != nil {
		log.Error(err)
		return
	}
	if err = databaseEngine.storeDecryptedGames(database["games"].(map[string]interface{})); err != nil {
		log.Error(err)
		return
	}
	if err = databaseEngine.storeDecryptedTools(database["win_tools"].(map[string]interface{})); err != nil {
		log.Error(err)
		return
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedConsoles(consolesJson map[string]interface{}) (err error) {
	for consoleKey, consoleValue := range consolesJson {
		var console *entity.Console
		if console, err = entity.ConsoleFromJSON(consoleKey, consoleValue); err != nil {
			return
		}
		log.Info("Storing " + console.Slug)
		databaseEngine.database.Create(console)
		consoleObject := consoleValue.(map[string]interface{})
		consoleFileTypesObject, _ := consoleObject["file_types"].(map[string]interface{})
		for actionKey, actionValue := range consoleFileTypesObject {
			for _, fileType := range actionValue.([]interface{}) {
				var consoleFileType *entity.ConsoleFileType
				if consoleFileType, err = entity.ConsoleFileTypeFromJSON(actionKey, console, fileType); err != nil {
					return
				}
				log.Info("Storing " + console.Slug + " " + consoleFileType.FileType + " file type")
				databaseEngine.database.Create(consoleFileType)
			}
		}
		for levelKey, levelValue := range consoleObject {
			if entity.ConsoleConfigIsLevel(levelKey) {
				consoleLevelObject := levelValue.(map[string]interface{})
				for consoleConfigName, consoleConfigValue := range consoleLevelObject {
					var consoleConfig *entity.ConsoleConfig
					if consoleConfig, err = entity.ConsoleConfigFromJSON(console, levelKey, consoleConfigName, consoleConfigValue); err != nil {
						return
					}
					log.Info("Storing " + console.Slug + " " + consoleConfig.Name + " configuration")
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
					var consoleLanguage *entity.ConsoleLanguage
					if consoleLanguage, err = entity.ConsoleLanguageFromJSON(console, uint(languageID), languageEntry); err != nil {
						return
					}
					log.Info("Storing " + console.Slug + " " + consoleLanguage.Name + " language")
					databaseEngine.database.Create(consoleLanguage)
				}
			}
		}
		if consolePluginsObject, ok := consoleObject["plugins"].(map[string]interface{}); ok {
			for pluginKey, pluginValue := range consolePluginsObject {
				var consolePlugin *entity.ConsolePlugin
				consolePlugin, err = entity.ConsolePluginFromJSON(pluginKey, console)
				log.Info("Storing " + console.Slug + " " + consolePlugin.Type + " plugin")
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
						var consolePluginsFile *entity.ConsolePluginsFile
						if consolePluginsFile, err = entity.ConsolePluginsFileFromJSON(
							consolePlugin, consolePluginCollectionPathValue,
							consolePluginDestinationValue,
							consolePluginFilesArray[fileIndex]); err != nil {
							return
						}
						log.Info("Storing " + console.Slug + " " + consolePlugin.Type + " plugin " + consolePluginsFile.Url + " file")
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
		var console entity.Console
		gameObject := gameValue.(map[string]interface{})
		if result := databaseEngine.database.First(&console, "slug = ?", gameObject["console_slug"].(string)); result.Error != nil {
			err = result.Error
			return
		}
		var game *entity.Game
		if game, err = entity.GameFromJSON(gameKey, &console, gameValue); err != nil {
			return
		}
		log.Info("Storing " + game.Slug + " game")
		databaseEngine.database.Create(game)

		collectionPath := gameObject["collection_path"]
		var gameDisks = []*entity.GameDisk{}
		if gameUrls, ok := gameObject["url"].([]interface{}); ok {
			for diskNumber := 0; diskNumber < len(gameUrls); diskNumber++ {
				var gameDisk *entity.GameDisk
				gameDiskImage := gameObject["disk_image"].([]interface{})[diskNumber]
				if gameDisk, err = entity.GameDiskFromJSON(game, uint(diskNumber), gameUrls[diskNumber], gameDiskImage, collectionPath); err != nil {
					return
				}
				gameDisks = append(gameDisks, gameDisk)
			}
		} else {
			var gameDisk *entity.GameDisk
			if gameDisk, err = entity.GameDiskFromJSON(game, 0, gameObject["url"], nil, collectionPath); err != nil {
				return
			}
			gameDisks = append(gameDisks, gameDisk)
		}
		for _, gameDisk := range gameDisks {
			log.Info("Storing " + game.Slug + " game disk " + strconv.Itoa(int(gameDisk.DiskNumber)))
			databaseEngine.database.Create(gameDisk)
		}

		if gameConfigObject, ok := gameObject["config"].(map[string]interface{}); ok {
			for configKey, configValue := range gameConfigObject {
				var gameConfig *entity.GameConfig
				if gameConfig, err = entity.GameConfigFromJSON(game, configKey, configValue); err != nil {
					return
				}
				log.Info("Storing " + game.Slug + " " + gameConfig.Name + " configuration")
				databaseEngine.database.Create(gameConfig)
			}
		}
		if gameAdditionlFilesObject, ok := gameObject["additional_files"].([]interface{}); ok {
			for _, gameAdditionlFileObject := range gameAdditionlFilesObject {
				var gameAdditionalFile *entity.GameAdditionalFile
				if gameAdditionalFile, err = entity.GameAdditionalFileFromJSON(game, gameAdditionlFileObject); err != nil {
					return
				}
				log.Info("Storing " + game.Slug + " " + gameAdditionalFile.Name + " additional File")
				databaseEngine.database.Create(gameAdditionalFile)
			}
		}
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedTools(toolsJson map[string]interface{}) (err error) {
	for toolKey, toolValue := range toolsJson {
		var tool *entity.Tool
		if tool, err = entity.ToolFromJSON(toolKey, toolsJson[toolKey]); err != nil {
			return
		}
		log.Info("Storing " + tool.Slug + " tool")
		databaseEngine.database.Create(tool)

		if toolFileTypesObject, ok := toolValue.(map[string]interface{})["file_types"].([]interface{}); ok {
			for _, toolFileTypeObject := range toolFileTypesObject {
				var toolFileType *entity.ToolFilesType
				if toolFileType, err = entity.ToolFilesTypeFromJSON(tool, toolFileTypeObject); err != nil {
					return
				}
				log.Info("Storing " + tool.Slug + "" + toolFileType.Tool.Url + " tool file")
				databaseEngine.database.Create(toolFileType)
			}
		}
	}
	return
}
