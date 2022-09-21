package storage

import (
	"os"
	"sync"

	"arkhive.dev/launcher/internal/folder"
)

type StorageEngine struct {
}

func NewStorageEngine() (instance *StorageEngine, err error) {
	instance = &StorageEngine{}
	return
}

func (storageEngine *StorageEngine) Initialize(waitGroup *sync.WaitGroup) {
	if _, err := os.Stat(folder.ROMS); os.IsNotExist(err) {
		if err = os.Mkdir(folder.TEMP, 0755); err != nil {
			panic(err)
		}
	}
	if _, err := os.Stat(folder.TEMP); os.IsNotExist(err) {
		if err = os.Mkdir(folder.TEMP, 0755); err != nil {
			panic(err)
		}
	}
}
