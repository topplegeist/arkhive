package database

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"strconv"

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

func NewDatabaseEngine() (instance *DatabaseEngine, err error) {
	instance = new(DatabaseEngine)
	instance.BootedEventEmitter = new(eventemitter.EventEmitter)
	instance.DecryptedEventEmitter = new(eventemitter.EventEmitter)

	go func() {
		log.Info("Connecting to database")
		if ok := instance.connectToDatabase(); !ok {
			panic("Cannot open database")
			return
		}
		log.Info("Applying database migrations")
		if err = instance.applyMigrations(); err != nil {
			panic(err)
			return
		}

		var storedDBHash []byte
		if storedDBHash, err = instance.getStoredDBHash(); err != nil {
			log.Fatal("Cannot decode the stored database hash")
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
		var dbData []byte

		if canDecrypt {
			if hashHasBeenStored {
				log.Info("Loading the encrypted database")
				if encryptedDBHash, err = instance.loadEncryptedDatabaseHash(); err != nil {
					panic(err)
				}
			}

			if !hashHasBeenStored || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
				if hashHasBeenStored {
					log.Info("The encrypted database hash not matches the one stored into the local database. Updating the local database")
				}
				log.Info("Loading encrypted database file")
				var encryptedDBData []byte
				if encryptedDBData, err = os.ReadFile(folder.CryptedDatabasePath); err != nil {
					log.Fatal("Cannot read the crypted database")
					panic(err)
				}
				log.Info("Loading database private key")
				var privateKey *rsa.PrivateKey
				if privateKey, err = instance.loadPrivateKey(); err != nil {
					panic(err)
				}
				log.Info("Decrypting encrypted database file")
				if dbData, err = encryption.Decrypt(privateKey, encryptedDBData); err != nil {
					log.Fatal("Cannot decode the encrypted database")
					panic(err)
				}

				if !hashHasBeenStored {
					log.Info("Calculating encrypted database file hash")
					hashEncoder := sha1.New()
					hashEncoder.Write(encryptedDBData)
					encryptedDBHash = hashEncoder.Sum(nil)
				}

				log.Info("Storing the database")
				if err = instance.storeDecryptedDatabase(dbData); err != nil {
					panic(err)
				}
				storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
				instance.setStoredDBHash(storingDBHash)
			} else {
				log.Info("No database updates")
			}
		} else if plainDbFileExists {
			log.Info("The encrypted database cannot be decrypted, proceeding with the plain JSON file")
			if dbData, err = os.ReadFile(folder.PlainDatabasePath); err != nil {
				log.Fatal("Cannot read the plain database file")
				panic(err)
			}

			if !keyFileExists {
				log.Info("The private key does not exists, generating a new key pair. It results in a new '" +
					folder.NewUndertowPath + "' file to be uploaded")
				var privateKey *rsa.PrivateKey
				if privateKey, err = encryption.GeneratePairKey(1024); err != nil {
					log.Fatal("Cannot generate the key pair")
					panic(err)
				}
				log.Info("Saving the private key file")
				privateKeyBytes := encryption.ExportPrivateKey(privateKey)
				if err = os.WriteFile(folder.DatabaseKeyPath, privateKeyBytes, 0644); err != nil {
					log.Fatal("Cannot write the private key file")
					panic(err)
				}
				var publicKeyBytes []byte
				log.Info("Saving the public key as " + folder.NewUndertowPath)
				if publicKeyBytes, err = encryption.ExportPublicKey(&privateKey.PublicKey); err != nil {
					log.Fatal("Cannot export the new undertow public key")
					panic(err)
				}
				if err = os.WriteFile(folder.NewUndertowPath, publicKeyBytes, 0644); err != nil {
					log.Fatal("Cannot write the temporary undertow file")
					panic(err)
				}
				if cryptedDbFileExists {
					log.Warn("The new key pair is different from the pair used to encrypt " + folder.CryptedDatabasePath +
						". arkHive will not delete the old " + folder.CryptedDatabasePath +
						" automatically. Please delete it before starting again the executable.")
				}
			}

			var encryptedDBData []byte
			if !cryptedDbFileExists {
				var privateKeyBytes []byte
				if privateKeyBytes, err = os.ReadFile(folder.DatabaseKeyPath); err != nil {
					log.Fatal("Cannot read the private key file")
					panic(err)
				}
				var privateKey *rsa.PrivateKey
				if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
					log.Fatal("Cannot import the private key")
					panic(err)
				}
				if encryptedDBData, err = encryption.Encrypt(&privateKey.PublicKey, dbData); err != nil {
					log.Fatal("Cannot encrypt the new encrypted database")
					panic(err)
				}
				if os.WriteFile(folder.CryptedDatabasePath, encryptedDBData, 0644); err != nil {
					log.Fatal("Cannot write the new encrypted database file")
					panic(err)
				}

				hashEncoder := sha1.New()
				hashEncoder.Write(encryptedDBData)
				encryptedDBHash = hashEncoder.Sum(nil)
			}

			if !hashHasBeenStored || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
				if err = instance.storeDecryptedDatabase(dbData); err != nil {
					panic(err)
				}
				storingDBHash := base64.URLEncoding.EncodeToString(encryptedDBHash)
				instance.setStoredDBHash(storingDBHash)
			}
		} else if !hashHasBeenStored {
			panic("No database to be imported")
		}

		instance.DecryptedEventEmitter.Emit(true)
		instance.BootedEventEmitter.Emit(true)
	}()
	return
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

func (databaseEngine *DatabaseEngine) loadEncryptedDatabaseHash() (encryptedDBHash []byte, err error) {
	var encryptedDBData []byte
	if encryptedDBData, err = os.ReadFile(folder.CryptedDatabasePath); err != nil {
		log.Fatal("Cannot read the encrypted database file")
		return
	}
	hashEncoder := sha1.New()
	hashEncoder.Write(encryptedDBData)
	encryptedDBHash = hashEncoder.Sum(nil)
	return
}

func (databaseEngine *DatabaseEngine) loadPrivateKey() (privateKey *rsa.PrivateKey, err error) {
	var privateKeyBytes []byte
	if privateKeyBytes, err = os.ReadFile(folder.DatabaseKeyPath); err != nil {
		log.Fatal("Cannot read the secret key file")
		return
	}
	if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
		log.Fatal("Cannot import the private key")
		return
	}
	return
}

func (databaseEngine DatabaseEngine) storeDecryptedDatabase(dbData []byte) (err error) {
	decoder := json.NewDecoder(bytes.NewReader(dbData))
	decoder.UseNumber()
	var database map[string]interface{}
	if err = decoder.Decode(&database); err != nil {
		log.Fatal(err)
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
			log.Info("Storing " + game.Slug + " game disk " + string(gameDisk.DiskNumber))
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
