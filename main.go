package main

import (
	"runtime/debug"
	"time"

	"arkhive.dev/launcher/common"
	"arkhive.dev/launcher/engines"
	log "github.com/sirupsen/logrus"
)

var stop = false

func stopMain(result bool) {
	stop = result
}

func main() {
	log.SetLevel(log.DebugLevel)
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatal("Failed to read build information")
		return
	}
	log.Debug("Launching arkHive v.", bi.Main.Version)

	databaseEngine, _ := engines.NewDatabaseEngine()

	databaseEngine.InitializationEndEventEmitter.Subscribe(stopMain)
	for !stop {
		time.Sleep(1 * time.Second)
	}
	log.Debug("Database initialized")
}
