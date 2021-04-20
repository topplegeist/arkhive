package storage

import (
	"os"

	"arkhive.dev/launcher/internal/engine/database"
	"arkhive.dev/launcher/internal/engine/network"
	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/eventemitter"
)

type StorageEngine struct {
	databaseEngine *database.DatabaseEngine
	networkEngine  *network.NetworkEngine

	// Event emitters
	BootedEventEmitter *eventemitter.EventEmitter
}

func NewStorageEngine(databaseEngine *database.DatabaseEngine, networkEngine *network.NetworkEngine) (instance *StorageEngine, err error) {
	instance = &StorageEngine{
		databaseEngine:     databaseEngine,
		networkEngine:      networkEngine,
		BootedEventEmitter: &eventemitter.EventEmitter{},
	}
	databaseEngine.DecryptedEventEmitter.Subscribe(instance.startEngine)
	return
}

func (storageEngine *StorageEngine) startEngine(_ bool) {
	if _, err := os.Stat(folder.ROMS); os.IsNotExist(err) {
		os.Mkdir(folder.TEMP, 0644)
	}
	if _, err := os.Stat(folder.TEMP); os.IsNotExist(err) {
		os.Mkdir(folder.TEMP, 0644)
	}
	storageEngine.BootedEventEmitter.Emit(true)
}
