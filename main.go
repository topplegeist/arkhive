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

var stop = false
var stop2 = false

func stopMain(result bool) {
	stop = result
}

func stopMain2(result bool) {
	stop2 = result
}

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

	databaseEngine, _ := engines.NewDatabaseEngine()
	networkEngine, _ := engines.NewNetworkEngine(databaseEngine, engines.GetUndertow())
	engines.NewSystemEngine(databaseEngine, networkEngine)
	engines.NewSearchEngine(databaseEngine)
	engines.NewStorageEngine(databaseEngine, networkEngine)
	engines.NewLauncherEngine(databaseEngine)

	databaseEngine.InitializationEndEventEmitter.Subscribe(stopMain)
	networkEngine.NetworkProcessInitializedEventEmitter.Subscribe(stopMain2)
	for !stop || !stop2 {
		time.Sleep(1 * time.Second)
	}
}
