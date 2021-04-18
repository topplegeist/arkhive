package engines

type SearchEngine struct {
	databaseEngine *DatabaseEngine
}

func NewSearchEngine(databaseEngine *DatabaseEngine) (instance *SearchEngine, err error) {
	instance = &SearchEngine{
		databaseEngine: databaseEngine,
	}
	return
}
