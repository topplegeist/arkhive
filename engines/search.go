package engines

import "arkhive.dev/launcher/common"

type SearchEngine struct {
	databaseEngine *DatabaseEngine

	// Event emitters
	BootedEventEmitter *common.EventEmitter
}

func NewSearchEngine(databaseEngine *DatabaseEngine) (instance *SearchEngine, err error) {
	instance = &SearchEngine{
		databaseEngine:     databaseEngine,
		BootedEventEmitter: &common.EventEmitter{},
	}
	go instance.BootedEventEmitter.Emit(true)
	return
}
