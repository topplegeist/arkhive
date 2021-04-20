package database

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"strconv"

	"arkhive.dev/launcher/internal/entity"
	"arkhive.dev/launcher/pkg/encryption"
	"arkhive.dev/launcher/pkg/eventemitter"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func NewDatabaseEngine() (instance *DatabaseEngine, err error) {
	instance = new(DatabaseEngine)
	instance.BootedEventEmitter = new(eventemitter.EventEmitter)
	instance.DecryptedEventEmitter = new(eventemitter.EventEmitter)

	go func() {
		if ok := instance.connectToDatabase(); !ok {
			log.Fatal("Cannot open database")
			return
		}
		if err = instance.applyMigrations(); err != nil {
			log.Fatal(err)
			return
		}

		var storedDBHash []byte
		if storedDBHashString := instance.getStoredDBHash(); storedDBHashString != "" {
			if storedDBHash, err = base64.URLEncoding.DecodeString(storedDBHashString); err != nil {
				log.Fatal("Cannot decode the stored database hash")
				log.Fatal(err)
				return
			}
		} else {
			log.Debug("Cannot get the stored database hash")
		}

		const cryptedDbFile = "db.honey"
		const plainDbFile = "db.json"
		const keyFilePath = "private_key.bee"
		const undertowPath = "undertow.tow"
		_, existenceFlag := os.Stat(cryptedDbFile)
		cryptedDbFileExists := !os.IsNotExist(existenceFlag)
		_, existenceFlag = os.Stat(plainDbFile)
		plainDbFileExists := !os.IsNotExist(existenceFlag)
		_, existenceFlag = os.Stat(keyFilePath)
		keyFileExists := !os.IsNotExist(existenceFlag)
		hashHasBeenCalculated := len(storedDBHash) > 0

		canDecrypt := cryptedDbFileExists && keyFileExists

		var encryptedDBData []byte
		var encryptedDBHash []byte
		var dbData []byte

		if canDecrypt {
			if hashHasBeenCalculated {
				if encryptedDBData, err = os.ReadFile(cryptedDbFile); err != nil {
					log.Fatal("Cannot read the encrypted database file")
					log.Fatal(err)
					return
				}
				hashEncoder := sha1.New()
				hashEncoder.Write(encryptedDBData)
				encryptedDBHash = hashEncoder.Sum(nil)
			}

			if !hashHasBeenCalculated || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
				if hashHasBeenCalculated {
					log.Info("The encrypted database hash not matches the one stored into the local database. Updating the local database.")
				}
				var privateKeyBytes []byte
				if privateKeyBytes, err = os.ReadFile(keyFilePath); err != nil {
					log.Fatal("Cannot read the secret key file")
					log.Fatal(err)
					return
				}
				var privateKey *rsa.PrivateKey
				if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
					log.Fatal("Cannot import the private key")
					log.Fatal(err)
					return
				}
				var encryptedDBData []byte
				if encryptedDBData, err = os.ReadFile(cryptedDbFile); err != nil {
					log.Fatal("Cannot read the crypted database")
					log.Fatal(err)
					return
				}
				if dbData, err = encryption.Decrypt(privateKey, encryptedDBData); err != nil {
					log.Fatal("Cannot decode the encrypted database")
					log.Fatal(err)
					return
				}

				if !hashHasBeenCalculated {
					hashEncoder := sha1.New()
					hashEncoder.Write(encryptedDBData)
					encryptedDBHash = hashEncoder.Sum(nil)
				}

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
				storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
				instance.setStoredDBHash(storingDBHash)
			}
		} else if plainDbFileExists {
			log.Info("The encrypted database cannot be decrypted, proceeding with the plain JSON file")
			if dbData, err = os.ReadFile(plainDbFile); err != nil {
				log.Fatal("Cannot read the plain database file")
				log.Fatal(err)
				return
			}

			if !keyFileExists {
				log.Info("The private key does not exists, generating a new key pair. It results in a new '" + undertowPath + "' file to be uploaded")
				var privateKey *rsa.PrivateKey
				if privateKey, err = encryption.GeneratePairKey(1024); err != nil {
					log.Fatal("Cannot generate the key pair")
					log.Fatal(err)
					return
				}
				privateKeyBytes := encryption.ExportPrivateKey(privateKey)
				if err = os.WriteFile(keyFilePath, privateKeyBytes, 0644); err != nil {
					log.Fatal("Cannot write the private key file")
					log.Fatal(err)
					return
				}
				var publicKeyBytes []byte
				if publicKeyBytes, err = encryption.ExportPublicKey(&privateKey.PublicKey); err != nil {
					log.Fatal("Cannot export the new undertow public key")
					log.Fatal(err)
					return
				}
				if err = os.WriteFile(undertowPath, publicKeyBytes, 0644); err != nil {
					log.Fatal("Cannot write the temporary undertow file")
					log.Fatal(err)
					return
				}
				if cryptedDbFileExists {
					log.Warn("The new key pair is different from the one used to encrypt " + cryptedDbFile + ". arkHive will not delete the old " + cryptedDbFile + " automatically. Please delete it before starting again the executable.")
				}
			}

			if !cryptedDbFileExists {
				var privateKeyBytes []byte
				if privateKeyBytes, err = os.ReadFile(keyFilePath); err != nil {
					log.Fatal("Cannot read the private key file")
					log.Fatal(err)
					return
				}
				var privateKey *rsa.PrivateKey
				if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
					log.Fatal("Cannot import the private key")
					log.Fatal(err)
					return
				}
				if encryptedDBData, err = encryption.Encrypt(&privateKey.PublicKey, dbData); err != nil {
					log.Fatal("Cannot encrypt the new encrypted database")
					log.Fatal(err)
					return
				}
				if os.WriteFile(cryptedDbFile, encryptedDBData, 0644); err != nil {
					log.Fatal("Cannot write the new encrypted database file")
					log.Fatal(err)
					return
				}
			}

			hashEncoder := sha1.New()
			hashEncoder.Write(encryptedDBData)
			encryptedDBHash = hashEncoder.Sum(nil)

			if !hashHasBeenCalculated || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
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
				storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
				instance.setStoredDBHash(storingDBHash)
			}
		} else if !hashHasBeenCalculated {
			panic("no database to be imported")
		}

		instance.DecryptedEventEmitter.Emit(true)
		instance.BootedEventEmitter.Emit(true)
	}()
	return
}

func (databaseEngine *DatabaseEngine) connectToDatabase() bool {
	const fileName string = "data.sqllite3"
	var err error
	databaseEngine.database, err = gorm.Open(sqlite.Open(fileName), &gorm.Config{
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

// User variable
func (databaseEngine DatabaseEngine) getStoredDBHash() string {
	var userVariable entity.UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "dbHash"); result.Error != nil || !userVariable.Value.Valid {
		return ""
	}
	return userVariable.Value.String
}

func (databaseEngine DatabaseEngine) setStoredDBHash(dbHash string) {
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

func (databaseEngine DatabaseEngine) GetLanguage() (Locale, error) {
	var userVariable entity.UserVariable
	if result := databaseEngine.database.First(&userVariable, "name = ?", "language"); result.Error != nil || !userVariable.Value.Valid {
		return ENGLISH, result.Error
	}
	language, err := strconv.Atoi(userVariable.Value.String)
	if err == nil {
		return ENGLISH, err
	}
	return Locale(language), nil
}

// Console
func (databaseEngine *DatabaseEngine) GetConsoles() (entity []entity.Console, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
}

func (databaseEngine *DatabaseEngine) GetConsoleByConsolePlugin(consolePlugin *entity.ConsolePlugin) (model entity.Console, err error) {
	err = databaseEngine.database.Model(consolePlugin).Association("Console").Find(&model)
	return
}

// Console Plugin
func (databaseEngine *DatabaseEngine) GetConsolePluginsByConsole(console *entity.Console) (entity []entity.ConsolePlugin, err error) {
	err = databaseEngine.database.Model(console).Association("ConsolePlugins").Find(&entity)
	return
}

// Console Plugin Files
func (databaseEngine *DatabaseEngine) GetConsolePluginsFilesByConsolePlugin(consolePlugin *entity.ConsolePlugin) (entity []entity.ConsolePluginsFile, err error) {
	err = databaseEngine.database.Model(consolePlugin).Association("ConsolePluginsFiles").Find(&entity)
	return
}

// Tool
func (databaseEngine *DatabaseEngine) GetTools() (entity []entity.Tool, err error) {
	if result := databaseEngine.database.Find(&entity); result.Error != nil {
		err = result.Error
		return
	}
	return
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
		var console *entity.Console
		if console, err = entity.ConsoleFromJSON(consoleKey, consoleValue); err != nil {
			return
		}
		databaseEngine.database.Create(console)
		consoleObject := consoleValue.(map[string]interface{})
		consoleFileTypesObject, _ := consoleObject["file_types"].(map[string]interface{})
		for actionKey, actionValue := range consoleFileTypesObject {
			for _, fileType := range actionValue.([]interface{}) {
				var consoleFileType *entity.ConsoleFileType
				if consoleFileType, err = entity.ConsoleFileTypeFromJSON(actionKey, console, fileType); err != nil {
					return
				}
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
					databaseEngine.database.Create(consoleLanguage)
				}
			}
		}
		if consolePluginsObject, ok := consoleObject["plugins"].(map[string]interface{}); ok {
			for pluginKey, pluginValue := range consolePluginsObject {
				var consolePlugin *entity.ConsolePlugin
				consolePlugin, err = entity.ConsolePluginFromJSON(pluginKey, console)
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
		databaseEngine.database.Create(game)
		collectionPath := gameObject["collection_path"]
		var gameDisk *entity.GameDisk
		if gameUrls, ok := gameObject["url"].([]interface{}); ok {
			for diskNumber := 0; diskNumber < len(gameUrls); diskNumber++ {
				gameDiskImage := gameObject["disk_image"].([]interface{})[diskNumber]
				if gameDisk, err = entity.GameDiskFromJSON(game, uint(diskNumber), gameUrls[diskNumber], gameDiskImage, collectionPath); err != nil {
					return
				}
				databaseEngine.database.Create(gameDisk)
			}
		} else {
			if gameDisk, err = entity.GameDiskFromJSON(game, 0, gameObject["url"], nil, collectionPath); err != nil {
				return
			}
			databaseEngine.database.Create(gameDisk)
		}
		if gameConfigObject, ok := gameObject["config"].(map[string]interface{}); ok {
			for configKey, configValue := range gameConfigObject {
				var gameConfig *entity.GameConfig
				if gameConfig, err = entity.GameConfigFromJSON(game, configKey, configValue); err != nil {
					return
				}
				databaseEngine.database.Create(gameConfig)
			}
		}
		if gameAdditionlFilesObject, ok := gameObject["additional_files"].([]interface{}); ok {
			for _, gameAdditionlFileObject := range gameAdditionlFilesObject {
				var gameAdditionalFile *entity.GameAdditionalFile
				if gameAdditionalFile, err = entity.GameAdditionalFileFromJSON(game, gameAdditionlFileObject); err != nil {
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
		var tool *entity.Tool
		if tool, err = entity.ToolFromJSON(toolKey, toolsJson[toolKey]); err != nil {
			return
		}
		databaseEngine.database.Create(tool)
		if toolFileTypesObject, ok := toolValue.(map[string]interface{})["file_types"].([]interface{}); ok {
			for _, toolFileTypeObject := range toolFileTypesObject {
				var toolFileType *entity.ToolFilesType
				if toolFileType, err = entity.ToolFilesTypeFromJSON(tool, toolFileTypeObject); err != nil {
					return
				}
				databaseEngine.database.Create(toolFileType)
			}
		}
	}
	return
}
