package database

import (
	"sync"

	"arkhive.dev/launcher/internal/database/delegate"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/sirupsen/logrus"
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
	if err = d.connectToDatabase(); err != nil {
		panic(err)
	}
	logrus.Info("Applying database migrations")
	if err = d.applyMigrations(); err != nil {
		panic(err)
	}

	// Check whether the database hash has been already saved on the database
	var storedDBHash []byte
	if storedDBHash, err = d.delegate.GetStoredDBHash(); err != nil {
		logrus.Error("Cannot decode the stored database hash")
		panic(err)
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
			panic(err)
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
		if err = d.delegate.StoreImported(importedConsoles, importedGames, importedTools); err != nil {
			logrus.Error(err)
		} else if err = d.delegate.SetStoredDBHash(encryptedDBHash); err != nil {
			panic(err)
		}
	}

	// End the routine
	waitGroup.Done()
}

func (d *Database) Deinitialize() {
	d.delegate.Close()
}

func (d *Database) connectToDatabase() error {
	return d.delegate.Open(d.basePath)
}

func (d Database) applyMigrations() (err error) {
	if err = d.delegate.Migrate(); err != nil {
		return err
	}
	return
}
