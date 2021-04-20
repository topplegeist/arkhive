package launcher

import (
	"arkhive.dev/launcher/internal/engine/database"
	"arkhive.dev/launcher/pkg/eventemitter"
)

type LauncherEngine struct {
	databaseEngine *database.DatabaseEngine

	// Event emitters
	BootedEventEmitter *eventemitter.EventEmitter
}

func NewLauncherEngine(databaseEngine *database.DatabaseEngine) (instance *LauncherEngine, err error) {
	instance = &LauncherEngine{
		databaseEngine:     databaseEngine,
		BootedEventEmitter: &eventemitter.EventEmitter{},
	}
	go instance.BootedEventEmitter.Emit(true)
	return
}
