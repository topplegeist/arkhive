package importer

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha1"
	"io"
	"reflect"

	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/encryption"
	log "github.com/sirupsen/logrus"
)

func ImportPlainDatabase(plainDatabaseReader io.Reader, databaseKeyReader io.Reader, databaseKeyWriter io.Writer,
	undertowWriter io.Writer, cryptedDatabaseWriter io.Writer,
	cryptedDatabaseAlreadyExists bool) (databaseData []byte, encryptedDBHash []byte, err error) {
	if databaseKeyReader != nil && databaseKeyWriter != nil {
		panic("private key reader and writer are both assigned")
	}
	if databaseKeyReader == nil && databaseKeyWriter == nil {
		panic("private key reader and writer are both not assigned")
	}

	if databaseKeyWriter != nil {
		log.Info("The private key does not exists, generating a new key pair. It results in a new '" +
			folder.NewUndertowPath + "' file to be uploaded")
		var privateKey *rsa.PrivateKey
		if privateKey, err = encryption.GeneratePairKey(1024); err != nil {
			log.Error("Cannot generate the key pair")
			panic(err)
		}
		log.Info("Saving the private key file")
		privateKeyBytes := encryption.ExportPrivateKey(privateKey)
		if _, err = databaseKeyWriter.Write(privateKeyBytes); err != nil {
			log.Error("Cannot write the private key file")
			panic(err)
		}
		var publicKeyBytes []byte
		log.Info("Saving the public key as " + folder.NewUndertowPath)
		if publicKeyBytes, err = encryption.ExportPublicKey(&privateKey.PublicKey); err != nil {
			log.Error("Cannot export the new undertow public key")
			panic(err)
		}
		if _, err = undertowWriter.Write(publicKeyBytes); err != nil {
			log.Error("Cannot write the temporary undertow file")
			panic(err)
		}
		if cryptedDatabaseAlreadyExists {
			log.Warn("The new key pair is different from the pair used to encrypt " + folder.CryptedDatabasePath +
				". arkHive will not delete the old " + folder.CryptedDatabasePath +
				" automatically. Please delete it before starting again the executable.")
		}
	} else {
		privateKeyBuffer := new(bytes.Buffer)
		if _, err = privateKeyBuffer.ReadFrom(databaseKeyReader); err != nil {
			log.Error("Cannot read the private key")
			panic(err)
		}
		var privateKey *rsa.PrivateKey
		if privateKey, err = encryption.ParsePrivateKey(privateKeyBuffer.Bytes()); err != nil {
			log.Error("Cannot import the private key")
			panic(err)
		}
		databaseBuffer := new(bytes.Buffer)
		if _, err = databaseBuffer.ReadFrom(plainDatabaseReader); err != nil {
			log.Error("Cannot read the plain database")
			panic(err)
		}
		databaseData = databaseBuffer.Bytes()

		var encryptedDBBytes []byte
		if encryptedDBBytes, err = encryption.Encrypt(&privateKey.PublicKey, databaseData); err != nil {
			log.Error("Cannot encrypt the new encrypted database")
			panic(err)
		}
		if _, err = cryptedDatabaseWriter.Write(encryptedDBBytes); err != nil {
			log.Error("Cannot write the new encrypted database file")
			panic(err)
		}

		hashEncoder := sha1.New()
		hashEncoder.Write(encryptedDBBytes)
		encryptedDBHash = hashEncoder.Sum(nil)
	}
	return
}

func ImportCryptedDatabase(cryptedDatabaseReader io.Reader, privateKey *rsa.PrivateKey, storedDBHash []byte) (databaseData []byte, encryptedDBHash []byte, err error) {
	log.Info("Loading the encrypted database")
	encryptedDBData := new(bytes.Buffer)
	if _, err = encryptedDBData.ReadFrom(cryptedDatabaseReader); err != nil {
		log.Error("Cannot read the encrypted database file")
		panic(err)
	}
	hashEncoder := sha1.New()
	hashEncoder.Write(encryptedDBData.Bytes())
	encryptedDBHash = hashEncoder.Sum(nil)

	if len(storedDBHash) == 0 || !reflect.DeepEqual(storedDBHash, encryptedDBHash) {
		if len(storedDBHash) > 0 {
			log.Info("The encrypted database hash not matches the one stored into the local database. Updating the local database")
		}
		log.Info("Decrypting encrypted database file")
		if databaseData, err = encryption.Decrypt(privateKey, encryptedDBData.Bytes()); err != nil {
			log.Error("Cannot decode the encrypted database")
			panic(err)
		}
	} else {
		log.Info("No database updates")
	}
	return
}
