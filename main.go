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
	if common.Debugging {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.ErrorLevel)
	}

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
	databaseEngine.BootedEventEmitter.Subscribe(func(_ bool) { databaseEngineStop = true })
	networkEngine, _ := engines.NewNetworkEngine(databaseEngine, engines.GetUndertow())
	networkEngine.BootedEventEmitter.Subscribe(func(_ bool) { networkEngineStop = true })
	systemEngine, _ := engines.NewSystemEngine(databaseEngine, networkEngine)
	systemEngine.BootedEventEmitter.Subscribe(func(_ bool) { systemEngineStop = true })
	searchEngine, _ := engines.NewSearchEngine(databaseEngine)
	searchEngine.BootedEventEmitter.Subscribe(func(_ bool) { searchEngineStop = true })
	storageEngine, _ := engines.NewStorageEngine(databaseEngine, networkEngine)
	storageEngine.BootedEventEmitter.Subscribe(func(_ bool) { storageEngineStop = true })
	launcherEngine, _ := engines.NewLauncherEngine(databaseEngine)
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
