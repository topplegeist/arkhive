package engines

type LauncherEngine struct {
	databaseEngine *DatabaseEngine
}

func NewLauncherEngine(databaseEngine *DatabaseEngine) (instance *LauncherEngine, err error) {
	instance = &LauncherEngine{
		databaseEngine: databaseEngine,
	}
	return
}
