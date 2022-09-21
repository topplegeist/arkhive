// Main library functionalities
package main

import (
	"flag"
	"runtime/debug"

	"arkhive.dev/launcher/internal/configloader"
	"arkhive.dev/launcher/internal/engine"
	"arkhive.dev/launcher/internal/gui"
	"arkhive.dev/launcher/internal/launcher"
	"arkhive.dev/launcher/internal/network"
	"arkhive.dev/launcher/internal/search"
	"arkhive.dev/launcher/internal/storage"
	"arkhive.dev/launcher/internal/system"
	"github.com/sirupsen/logrus"
)

// Name of the current application. Used to load the configuration.
const APPLICATION_NAME = "arkhive"

type ApplicationEngineTypes int8

const (
	Database ApplicationEngineTypes = iota
	Network
	System
	Search
	Storage
	Launcher
	EnginesCount
)

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

	setLogger(configuration.LogLevel)

	if *configurationFilePath != "" {
		logrus.Infof("Loaded config file %s", *configurationFilePath)
	}

	logrus.Infof("Log level set to %s", configuration.LogLevel)

	logBuildInformation()

	runEngines(configuration)

	//databaseEngineStop := false
	//networkEngineStop := false
	//systemEngineStop := false
	//searchEngineStop := false
	//storageEngineStop := false
	//launcherEngineStop := false
	//
	//databaseEngine := database.NewDatabaseEngine()
	//databaseEngine.BootedEventEmitter.Subscribe(func(_ bool) { databaseEngineStop = true })
	//go databaseEngine.Initialize()
	//networkEngine, _ := network.NewNetworkEngine(databaseEngine, system.GetUndertow())
	//networkEngine.BootedEventEmitter.Subscribe(func(_ bool) { networkEngineStop = true })
	//systemEngine, _ := system.NewSystemEngine(databaseEngine, networkEngine)
	//systemEngine.BootedEventEmitter.Subscribe(func(_ bool) { systemEngineStop = true })
	//searchEngine, _ := search.NewSearchEngine(databaseEngine)
	//searchEngine.BootedEventEmitter.Subscribe(func(_ bool) { searchEngineStop = true })
	//storageEngine, _ := storage.NewStorageEngine(databaseEngine, networkEngine)
	//storageEngine.BootedEventEmitter.Subscribe(func(_ bool) { storageEngineStop = true })
	//launcherEngine, _ := launcher.NewLauncherEngine(databaseEngine)
	//launcherEngine.BootedEventEmitter.Subscribe(func(_ bool) { launcherEngineStop = true })

	//for !databaseEngineStop ||
	//	!networkEngineStop ||
	//	!systemEngineStop ||
	//	!searchEngineStop ||
	//	!storageEngineStop ||
	//	!launcherEngineStop {
	//	time.Sleep(1 * time.Second)
	//}
}

// Parse and set the log level
func setLogger(levelString string) {
	level, err := logrus.ParseLevel(levelString)
	if err != nil {
		logrus.Errorf("%+v", err)
		return
	}
	logrus.SetLevel(level)
}

// Display build information
func logBuildInformation() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		panic("Failed to read build information")
	}
	logrus.Debugf("Launching arkHive v.%s", bi.Main.Version)
}

// Run the core application engines
func runEngines(configuration configloader.Config) {
	var engines []engine.ApplicationEngine = make([]engine.ApplicationEngine, EnginesCount)
	// The application entities data
	//engines[Database] = database.NewDatabaseEngine(configuration.BasePath)
	// The handler of the communication
	engines[Network], _ = network.NewNetworkEngine()
	// The operative systems and hardware adapter
	engines[System], _ = system.NewSystemEngine()
	// The data scraper
	engines[Search], _ = search.NewSearchEngine()
	// The engine to persist large amount of unscrepable
	engines[Storage], _ = storage.NewStorageEngine()
	// The interface with the emulation side
	engines[Launcher], _ = launcher.NewLauncherEngine()

	guiHandler := gui.QtHandler{}

	engineController := engine.NewController(engines, &guiHandler)
	engineController.Initialize()
}
