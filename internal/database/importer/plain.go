package importer

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"os"
	"path/filepath"
	"reflect"

	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/encryption"
	"github.com/sirupsen/logrus"
)

type PlainImporter struct {
	basePath string
	Consoles []Console
	Games    []Game
	Tools    []Tool
}

func (p *PlainImporter) Import(currentDBHash []byte) (importedDBHash []byte, err error) {
	var databaseData []byte
	if p.CanImport() {
		databaseData, importedDBHash, err = p.load(currentDBHash)
	}
	return
}

func (p *PlainImporter) CanImport() bool {
	// Check if a plain database file and the key file exists
	logrus.Debug("Checking if a plain database could be imported")
	_, existenceFlag := os.Stat(filepath.Join(p.basePath, folder.PlainDatabasePath))
	canImport := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(filepath.Join(p.basePath, folder.DatabaseKeyPath))
	keyFileExists := !os.IsNotExist(existenceFlag)
	if !canImport {
		logrus.Debug("The plain database is not present")
	}
	if !keyFileExists {
		logrus.Debug("Cannot load the private key")
	}
	return canImport && keyFileExists
}

func (p *PlainImporter) load(currentDBHash []byte) (databaseData []byte, encryptedDBHash []byte, err error) {
	// Read the database file to be imported
	var plainDatabaseFileReader *os.File
	if plainDatabaseFileReader, err = os.Open(filepath.Join(p.basePath, folder.PlainDatabasePath)); err != nil {
		logrus.Error("Cannot read the plain database file")
		panic(err)
	}
	defer plainDatabaseFileReader.Close()
	databaseBuffer := &bytes.Buffer{}
	if _, err = databaseBuffer.ReadFrom(plainDatabaseFileReader); err != nil {
		logrus.Error("Cannot read the plain database")
		panic(err)
	}
	databaseData = databaseBuffer.Bytes()

	// Import the key file
	logrus.Info("Importing the database key file")
	var privateKey *rsa.PrivateKey
	if privateKey, err = encryption.ParsePrivateKeyFile(filepath.Join(p.basePath, folder.DatabaseKeyPath)); err != nil {
		logrus.Error("Cannot parse the private key file")
		panic(err)
	}

	// Write the encrypted database file or override it if already exists
	logrus.Info("Encrypting the database")
	var encryptedDBBytes []byte
	if encryptedDBBytes, err = encryption.Encrypt(&privateKey.PublicKey, databaseData); err != nil {
		logrus.Error("Cannot encrypt the new encrypted database")
		panic(err)
	}

	// Check if exists a copy of the encrypted database
	_, existenceFlag := os.Stat(filepath.Join(p.basePath, folder.EncryptedDatabasePath))
	if !os.IsNotExist(existenceFlag) {
		os.Remove(filepath.Join(p.basePath, folder.EncryptedDatabasePath))
	}
	var encryptedDatabaseWriter *os.File
	if encryptedDatabaseWriter, err = os.Create(filepath.Join(p.basePath, folder.EncryptedDatabasePath)); err != nil {
		logrus.Error("Cannot create the encrypted database file")
		panic(err)
	}
	defer encryptedDatabaseWriter.Close()
	if _, err = encryptedDatabaseWriter.Write(encryptedDBBytes); err != nil {
		logrus.Error("Cannot write the new encrypted database file")
		panic(err)
	}

	logrus.Info("Calculating the database hash")
	// Calculate the hash of the new encrypted database
	hashEncoder := sha1.New()
	hashEncoder.Write(encryptedDBBytes)
	encryptedDBHash = hashEncoder.Sum(nil)

	// Return the database file if the database has never been imported and if the hash stored in the database is different from that taken from the current file
	if len(currentDBHash) == 0 || !reflect.DeepEqual(currentDBHash, encryptedDBHash) {
		logrus.Info("The encrypted database hash does not match the one stored into the local database. Updating the local database")
	} else {
		logrus.Info("No database updates")
		databaseData = nil
	}

	return
}

func (p PlainImporter) GetConsoles() (consoles []Console) {
	return p.Consoles
}

func (p PlainImporter) GetGames() (games []Game) {
	return p.Games
}

func (p PlainImporter) GetTools() (tools []Tool) {
	return p.Tools
}
