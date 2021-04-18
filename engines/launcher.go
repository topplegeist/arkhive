package engines

import "arkhive.dev/launcher/common"

type LauncherEngine struct {
	databaseEngine *DatabaseEngine

	// Event emitters
	BootedEventEmitter *common.EventEmitter
}

func NewLauncherEngine(databaseEngine *DatabaseEngine) (instance *LauncherEngine, err error) {
	instance = &LauncherEngine{
		databaseEngine:     databaseEngine,
		BootedEventEmitter: &common.EventEmitter{},
	}
	go instance.BootedEventEmitter.Emit(true)
	return
}
