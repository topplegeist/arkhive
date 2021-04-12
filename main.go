package main

import (
	"runtime/debug"
	"time"

	"arkhive.dev/launcher/engines"
	"arkhive.dev/launcher/models/network"
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

	databaseEngine, _ := engines.NewDatabaseEngine()
	networkEngine, _ := engines.NewNetworkEngine(databaseEngine, network.GetUndertow())

	databaseEngine.InitializationEndEventEmitter.Subscribe(stopMain)
	networkEngine.NetworkProcessInitialized.Subscribe(stopMain2)
	for !stop || !stop2 {
		time.Sleep(1 * time.Second)
	}
}
