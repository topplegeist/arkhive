package importer

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"arkhive.dev/launcher/pkg/encryption"
	"github.com/sirupsen/logrus"
)

const EncryptedDatabasePath = "db.honey"
const DatabaseKeyPath = "private_key.bee"

type EncryptedImporter struct {
	Plain    Plain
	basePath string
}

func NewEncryptedImporter(basePath string) *EncryptedImporter {
	return &EncryptedImporter{
		Plain{
			basePath: basePath,
			consoles: []Console{},
			games:    []Game{},
			tools:    []Tool{},
		},
		basePath,
	}
}

func (e *EncryptedImporter) Import(currentDBHash []byte) (importedDBHash []byte, err error) {
	if !e.canLoad() {
		logrus.Debug("Cannot load the encrypted database")
		return nil, nil
	}

	// Load the user private key used to encrypt the database
	logrus.Info("Loading database private key")
	var privateKeyBytes []byte
	if privateKeyBytes, err = os.ReadFile(filepath.Join(e.basePath, DatabaseKeyPath)); err != nil {
		logrus.Error("Cannot read the secret key file")
		return
	}
	var privateKey *rsa.PrivateKey
	if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
		logrus.Error("Cannot import the private key")
		return
	}

	// Load the encrypted database file
	var encryptedDatabaseReader io.Reader
	if encryptedDatabaseReader, err = os.Open(filepath.Join(e.basePath, EncryptedDatabasePath)); err != nil {
		logrus.Error("Cannot read the database key file")
		return
	}
	logrus.Info("Loading the encrypted database")
	encryptedDBData := &bytes.Buffer{}
	if _, err = encryptedDBData.ReadFrom(encryptedDatabaseReader); err != nil {
		logrus.Error("Cannot read the encrypted database file")
		return
	}

	// Calculate the encrypted database file hash
	hashEncoder := sha1.New()
	if _, err = hashEncoder.Write(encryptedDBData.Bytes()); err != nil {
		return
	}
	encryptedDBHash := hashEncoder.Sum(nil)

	// Return the database file if the database has never been imported and if the hash stored in the database is different from that of the current file
	if !reflect.DeepEqual(currentDBHash, encryptedDBHash) {
		if len(currentDBHash) > 0 {
			logrus.Info("The encrypted database hash not matches the one stored into the local database. Updating the local database")
		}
		logrus.Info("Decrypting encrypted database file")
		var databaseData []byte
		if databaseData, err = encryption.Decrypt(privateKey, encryptedDBData.Bytes()); err != nil {
			logrus.Error("Cannot decode the encrypted database")
			return
		}
		var plainDatabaseFile *os.File
		if plainDatabaseFile, err = os.OpenFile(filepath.Join(e.basePath, PlainDatabasePath), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
			return
		}
		defer plainDatabaseFile.Close()
		if _, err = plainDatabaseFile.Write(databaseData); err != nil {
			return
		}
		if _, err = e.Plain.Import(currentDBHash); err != nil {
			return
		}
		importedDBHash = encryptedDBHash
	} else {
		logrus.Info("No database updates")
	}
	return
}

func (e *EncryptedImporter) canLoad() bool {
	// Check if both the encrypted database file and the user private key exists
	logrus.Debug("Checking if an encrypted database could be imported")
	_, existenceFlag := os.Stat(filepath.Join(e.basePath, EncryptedDatabasePath))
	encryptedDbFileExists := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(filepath.Join(e.basePath, DatabaseKeyPath))
	keyFileExists := !os.IsNotExist(existenceFlag)
	return encryptedDbFileExists && keyFileExists
}

func (e *EncryptedImporter) GetConsoles() []Console {
	return e.Plain.GetConsoles()
}
func (e *EncryptedImporter) GetGames() []Game {
	return e.Plain.GetGames()
}
func (e *EncryptedImporter) GetTools() []Tool {
	return e.Plain.GetTools()
}
