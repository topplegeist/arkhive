package main

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"arkhive.dev/launcher/internal/engine/database"
	"arkhive.dev/launcher/internal/engine/launcher"
	"arkhive.dev/launcher/internal/engine/network"
	"arkhive.dev/launcher/internal/engine/search"
	"arkhive.dev/launcher/internal/engine/storage"
	"arkhive.dev/launcher/internal/engine/system"
	"arkhive.dev/launcher/internal/environment"
	log "github.com/sirupsen/logrus"
)

func main() {
	if environment.Debugging {
		log.SetLevel(log.DebugLevel)
		log.SetReportCaller(true)
	} else {
		log.SetLevel(log.ErrorLevel)
	}
	log.SetFormatter(&log.TextFormatter{})

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build information")
	}
	log.Debug("Launching arkHive v.", bi.Main.Version)

	if environment.Debugging {
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

	databaseEngine, _ := database.NewDatabaseEngine()
	databaseEngine.BootedEventEmitter.Subscribe(func(_ bool) { databaseEngineStop = true })
	networkEngine, _ := network.NewNetworkEngine(databaseEngine, system.GetUndertow())
	networkEngine.BootedEventEmitter.Subscribe(func(_ bool) { networkEngineStop = true })
	systemEngine, _ := system.NewSystemEngine(databaseEngine, networkEngine)
	systemEngine.BootedEventEmitter.Subscribe(func(_ bool) { systemEngineStop = true })
	searchEngine, _ := search.NewSearchEngine(databaseEngine)
	searchEngine.BootedEventEmitter.Subscribe(func(_ bool) { searchEngineStop = true })
	storageEngine, _ := storage.NewStorageEngine(databaseEngine, networkEngine)
	storageEngine.BootedEventEmitter.Subscribe(func(_ bool) { storageEngineStop = true })
	launcherEngine, _ := launcher.NewLauncherEngine(databaseEngine)
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
