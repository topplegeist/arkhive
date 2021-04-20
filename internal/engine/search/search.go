package search

import (
	"arkhive.dev/launcher/internal/engine/database"
	"arkhive.dev/launcher/pkg/eventemitter"
)

type SearchEngine struct {
	databaseEngine *database.DatabaseEngine

	// Event emitters
	BootedEventEmitter *eventemitter.EventEmitter
}

func NewSearchEngine(databaseEngine *database.DatabaseEngine) (instance *SearchEngine, err error) {
	instance = &SearchEngine{
		databaseEngine:     databaseEngine,
		BootedEventEmitter: &eventemitter.EventEmitter{},
	}
	go instance.BootedEventEmitter.Emit(true)
	return
}
