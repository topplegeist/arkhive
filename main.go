package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"arkhive.dev/launcher/common"
	"arkhive.dev/launcher/engines"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build information")
	}
	log.Debug("Launching arkHive v.", bi.Main.Version)
	if common.Debugging {
		debuggingPath := filepath.Join("..", "..", "build", "go")
		if _, err := os.Stat(debuggingPath); os.IsNotExist(err) {
			os.Mkdir(debuggingPath, 0644)
		}
		os.Chdir(debuggingPath)
	}

	databaseEngineStop := false
	networkEngineStop := false
	systemEngineStop := false
	searchEngineStop := false
	storageEngineStop := false
	launcherEngineStop := false

	databaseEngine, _ := engines.NewDatabaseEngine()
	networkEngine, _ := engines.NewNetworkEngine(databaseEngine, engines.GetUndertow())
	systemEngine, _ := engines.NewSystemEngine(databaseEngine, networkEngine)
	searchEngine, _ := engines.NewSearchEngine(databaseEngine)
	storageEngine, _ := engines.NewStorageEngine(databaseEngine, networkEngine)
	launcherEngine, _ := engines.NewLauncherEngine(databaseEngine)

	databaseEngine.BootedEventEmitter.Subscribe(func(_ bool) { databaseEngineStop = true })
	networkEngine.BootedEventEmitter.Subscribe(func(_ bool) { networkEngineStop = true })
	systemEngine.BootedEventEmitter.Subscribe(func(_ bool) { systemEngineStop = true })
	searchEngine.BootedEventEmitter.Subscribe(func(_ bool) { searchEngineStop = true })
	storageEngine.BootedEventEmitter.Subscribe(func(_ bool) { storageEngineStop = true })
	launcherEngine.BootedEventEmitter.Subscribe(func(_ bool) { launcherEngineStop = true })

	for !databaseEngineStop ||
		!networkEngineStop ||
		!systemEngineStop ||
		!searchEngineStop ||
		!storageEngineStop ||
		!launcherEngineStop {
		time.Sleep(1 * time.Second)
	}
}
