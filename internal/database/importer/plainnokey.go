package importer

import (
	"bytes"
	"crypto/rsa"
	"os"
	"path/filepath"

	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/encryption"
	"github.com/sirupsen/logrus"
)

type PlainNoKeyImporter struct {
	basePath string
	//currentDBHash []byte
}

func (plainNoKeyImporter *PlainNoKeyImporter) CanImport() bool {
	// Check if a plain database file exists but the key file not exists
	logrus.Debug("Checking if a plain database could be imported")
	_, existenceFlag := os.Stat(filepath.Join(plainNoKeyImporter.basePath, folder.PlainDatabasePath))
	canImport := !os.IsNotExist(existenceFlag)
	_, existenceFlag = os.Stat(filepath.Join(plainNoKeyImporter.basePath, folder.DatabaseKeyPath))
	keyFileExists := !os.IsNotExist(existenceFlag)
	if !canImport {
		logrus.Debug("The plain database is not present")
	}
	if keyFileExists {
		logrus.Debug("Can load the private key")
	}
	return canImport && !keyFileExists
}

func (plainNoKeyImporter *PlainNoKeyImporter) Import() (databaseData []byte, encryptedDBHash []byte, err error) {
	// Read the database file
	var plainDatabaseFileReader *os.File
	if plainDatabaseFileReader, err = os.Open(filepath.Join(plainNoKeyImporter.basePath, folder.PlainDatabasePath)); err != nil {
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

	// Check if the user key file exists
	// Check if exists a copy of the encrypted database
	_, existenceFlag := os.Stat(filepath.Join(plainNoKeyImporter.basePath, folder.EncryptedDatabasePath))
	encryptedDbFileExists := !os.IsNotExist(existenceFlag)

	// Generate and save the new private key
	logrus.Infof("The private key does not exists, generating a new key pair. It results in a new '%s' file to be uploaded", folder.NewUndertowPath)
	var privateKey *rsa.PrivateKey
	if privateKey, err = encryption.GeneratePairKey(1024); err != nil {
		logrus.Error("Cannot generate the key pair")
		panic(err)
	}
	privateKeyBytes := encryption.ExportPrivateKey(privateKey)
	logrus.Info("Saving the private key file")
	var databaseKeyWriter *os.File
	if databaseKeyWriter, err = os.Create(filepath.Join(plainNoKeyImporter.basePath, folder.DatabaseKeyPath)); err != nil {
		logrus.Error("Cannot create the database key file")
		panic(err)
	}
	defer databaseKeyWriter.Close()
	if _, err = databaseKeyWriter.Write(privateKeyBytes); err != nil {
		logrus.Error("Cannot write the private key file")
		panic(err)
	}

	// Save the new public key as the new undertow
	var publicKeyBytes []byte
	logrus.Infof("Saving the public key as %s", folder.NewUndertowPath)
	if publicKeyBytes, err = encryption.ExportPublicKey(&privateKey.PublicKey); err != nil {
		logrus.Error("Cannot export the new undertow public key")
		panic(err)
	}
	var undertowWriter *os.File
	if undertowWriter, err = os.Create(filepath.Join(plainNoKeyImporter.basePath, folder.NewUndertowPath)); err != nil {
		logrus.Error("Cannot create the new undertow file")
		panic(err)
	}
	defer undertowWriter.Close()
	if _, err = undertowWriter.Write(publicKeyBytes); err != nil {
		logrus.Error("Cannot write the temporary undertow file")
		panic(err)
	}
	if encryptedDbFileExists {
		logrus.Warnf("The new key pair is different from the pair used to encrypt %s. "+
			"arkHive will not delete the old %s automatically. "+
			"Please delete it before starting again the executable.",
			folder.EncryptedDatabasePath, folder.EncryptedDatabasePath)
	}

	return
}
