package importer

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/encryption"
	"github.com/sirupsen/logrus"
)

type EncryptedImporter struct {
	basePath      string
	currentDBHash []byte
}

func (encryptedImporter *EncryptedImporter) CanImport() (canImport bool) {
	// Check if both the encrypted database file and the user private key exists
	logrus.Debug("Checking if an encrypted database could be imported")
	_, existenceFlag := os.Stat(filepath.Join(encryptedImporter.basePath, folder.EncryptedDatabasePath))
	encryptedDbFileExists := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(filepath.Join(encryptedImporter.basePath, folder.DatabaseKeyPath))
	keyFileExists := !os.IsNotExist(existenceFlag)
	canImport = encryptedDbFileExists && keyFileExists
	if !canImport {
		logrus.Debugf("The encrypted database could be imported (database: %t, key: %t)", encryptedDbFileExists, keyFileExists)
	}
	return
}

func (encryptedImporter *EncryptedImporter) Import() (databaseData []byte, encryptedDBHash []byte, err error) {
	// Load the user private key used to encrypt the database
	logrus.Info("Loading database private key")
	var privateKeyBytes []byte
	if privateKeyBytes, err = os.ReadFile(filepath.Join(encryptedImporter.basePath, folder.DatabaseKeyPath)); err != nil {
		logrus.Error("Cannot read the secret key file")
		panic(err)
	}
	var privateKey *rsa.PrivateKey
	if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
		logrus.Error("Cannot import the private key")
		panic(err)
	}

	// Load the encrypted database file
	var encryptedDatabaseReader io.Reader
	if encryptedDatabaseReader, err = os.Open(filepath.Join(encryptedImporter.basePath, folder.EncryptedDatabasePath)); err != nil {
		logrus.Error("Cannot read the database key file")
		panic(err)
	}
	logrus.Info("Loading the encrypted database")
	encryptedDBData := &bytes.Buffer{}
	if _, err = encryptedDBData.ReadFrom(encryptedDatabaseReader); err != nil {
		logrus.Error("Cannot read the encrypted database file")
		panic(err)
	}

	// Calculate the encrypted database file hash
	hashEncoder := sha1.New()
	hashEncoder.Write(encryptedDBData.Bytes())
	encryptedDBHash = hashEncoder.Sum(nil)

	// Return the database file if the database has never been imported and if the hash stored in the database is different from that of the current file
	if len(encryptedImporter.currentDBHash) == 0 || !reflect.DeepEqual(encryptedImporter.currentDBHash, encryptedDBHash) {
		if len(encryptedImporter.currentDBHash) > 0 {
			logrus.Info("The encrypted database hash not matches the one stored into the local database. Updating the local database")
		}
		logrus.Info("Decrypting encrypted database file")
		if databaseData, err = encryption.Decrypt(privateKey, encryptedDBData.Bytes()); err != nil {
			logrus.Error("Cannot decode the encrypted database")
			panic(err)
		}
	} else {
		logrus.Info("No database updates")
	}
	return
}
