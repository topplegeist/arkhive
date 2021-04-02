package engines

import (
	"crypto/sha1"
	"encoding/base64"
	"os"
	"reflect"

	"arkhive.dev/launcher/models"
	log "github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseEngine struct {
	database *gorm.DB
}

func NewDatabaseEngine() (instance *DatabaseEngine, err error) {
	instance = new(DatabaseEngine)
	if ok := instance.connectToDatabase(); !ok {
		log.Fatal("Cannot open database")
		return
	}
	instance.applyMigrations()
	storedEncryptedDBHashString := instance.getStoredEncryptedDBHash()
	var storedEncryptedDBHash []byte
	if storedEncryptedDBHashString != "" {
		if storedEncryptedDBHash, err = base64.URLEncoding.DecodeString(storedEncryptedDBHashString); err != nil {
			log.Fatal("Cannot decode the stored encrypted database hash")
			return
		}
	} else {
		log.Debug("Cannot get the stored encrypted database hash")
	}

	cryptedDbFile := "db.bee"
	keyFilePath := "private_key.bee"
	_, existenceFlag := os.Stat(cryptedDbFile)
	cryptedDbFileExists := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(keyFilePath)
	keyFileExists := !os.IsNotExist(existenceFlag)

	canDecrypt := cryptedDbFileExists && keyFileExists
	if canDecrypt {
		var encryptedDBData []byte
		if encryptedDBData, err = os.ReadFile(cryptedDbFile); err != nil {
			log.Fatal(err)
			return
		}
		hashEncoder := sha1.New()
		hashEncoder.Write(encryptedDBData)
		encryptedDBHash := hashEncoder.Sum(nil)

		if !reflect.DeepEqual(storedEncryptedDBHash, encryptedDBHash) {
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
	}
	return
}

func (databaseEngine *DatabaseEngine) connectToDatabase() bool {
	const fileName string = "data.sqllite3"
	var err error
	databaseEngine.database, err = gorm.Open(sqlite.Open(fileName), &gorm.Config{})
	return err == nil
}

func (databaseEngine DatabaseEngine) applyMigrations() {
	databaseEngine.database.AutoMigrate(&models.User{},
		&models.Chat{}, &models.Tool{}, &models.Console{}, &models.Game{},
		&models.ToolFileType{}, &models.ConsoleFileType{}, &models.ConsoleLanguage{},
		&models.ConsolePlugin{}, &models.ConsolePluginsFile{},
		&models.ConsoleConfig{}, &models.GameDisk{}, &models.GameAdditionalFile{},
		&models.GameConfig{}, &models.UserVariable{})
}

func (databaseEngine DatabaseEngine) getStoredEncryptedDBHash() string {
	var userVariable models.UserVariable
	if result := databaseEngine.database.First(&userVariable, "dbHash"); result.Error != nil || !userVariable.Value.Valid {
		return ""
	}
	return userVariable.Value.String
}
