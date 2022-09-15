package main

import (
	"flag"
	"runtime/debug"
	"time"

	"arkhive.dev/launcher/internal/configloader"
	"arkhive.dev/launcher/internal/engine/database"
	"arkhive.dev/launcher/internal/engine/launcher"
	"arkhive.dev/launcher/internal/engine/network"
	"arkhive.dev/launcher/internal/engine/search"
	"arkhive.dev/launcher/internal/engine/storage"
	"arkhive.dev/launcher/internal/engine/system"
	"github.com/sirupsen/logrus"
)

// Name of the current application. Used to load the configuration.
const APPLICATION_NAME = "arkhive"

func main() {
	// Parsing the command line argument to change settings file location
	configurationFilePath := flag.String("config", "", "Configuration file path")
	flag.Parse()
	// Loading application configuration
	configuration, err := configloader.LoadConfiguration(APPLICATION_NAME, *configurationFilePath)
	if err != nil {
		logrus.Errorf("%+v", err)
		return
	}
	level, err := logrus.ParseLevel(configuration.LogLevel)
	if err != nil {
		logrus.Errorf("%+v", err)
		return
	}

	// Set log level
	logrus.SetLevel(level)
	if *configurationFilePath != "" {
		logrus.Infof("Loaded config file %s", *configurationFilePath)
	}
	logrus.Infof("Setting log level to %s", level.String())

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build information")
	}
	logrus.Debug("Launching arkHive v.", bi.Main.Version)

	databaseEngineStop := false
	networkEngineStop := false
	systemEngineStop := false
	searchEngineStop := false
	storageEngineStop := false
	launcherEngineStop := false

	databaseEngine := database.NewDatabaseEngine()
	databaseEngine.BootedEventEmitter.Subscribe(func(_ bool) { databaseEngineStop = true })
	go databaseEngine.Initialize()
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
