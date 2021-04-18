package engines

import (
	"os"

	"arkhive.dev/launcher/common"
)

type StorageEngine struct {
	databaseEngine *DatabaseEngine
	networkEngine  *NetworkEngine

	// Event emitters
	BootedEventEmitter *common.EventEmitter
}

func NewStorageEngine(databaseEngine *DatabaseEngine, networkEngine *NetworkEngine) (instance *StorageEngine, err error) {
	instance = &StorageEngine{
		databaseEngine:     databaseEngine,
		networkEngine:      networkEngine,
		BootedEventEmitter: &common.EventEmitter{},
	}
	databaseEngine.DecryptedEventEmitter.Subscribe(instance.startEngine)
	return
}

func (storageEngine *StorageEngine) startEngine(_ bool) {
	if _, err := os.Stat(common.STORAGE_ROMS_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.TEMP_DOWNLOAD_FOLDER_PATH, 0644)
	}
	if _, err := os.Stat(common.TEMP_DOWNLOAD_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.TEMP_DOWNLOAD_FOLDER_PATH, 0644)
	}
	storageEngine.BootedEventEmitter.Emit(true)
}
