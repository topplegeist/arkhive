package engine

import (
	"database/sql"
	"encoding/base64"
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
	var encryptedDBHash []byte
	var (
		importedConsoles []importer.Console
		importedGames    []importer.Game
		importedTools    []importer.Tool
	)
	for _, importer := range d.importers {
		encryptedDBHash, err = importer.Import(storedDBHash)
		if err != nil {
			panic("error importing the database")
		}
		if encryptedDBHash != nil {
			importedConsoles = importer.GetConsoles()
			importedGames = importer.GetGames()
			importedTools = importer.GetTools()
			break
		}
	}

	// Parse the database data read, if any
	if encryptedDBHash != nil {
		logrus.Info("Storing the new imported database")
		if err = d.storeImportedDatabase(importedConsoles, importedGames, importedTools); err != nil {
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

func (d Database) storeImportedDatabase(consoles []importer.Console, games []importer.Game, tools []importer.Tool) (err error) {
	/*

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
		}*/
	return
}

func (d Database) storeDecryptedConsoles(consolesJson map[string]interface{}) (err error) {
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
