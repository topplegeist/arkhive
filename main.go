package main

import (
	"runtime/debug"

	"arkhive.dev/launcher/engines"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Fatal("Failed to read build information")
		return
	}
	log.Debug("Launching arkHive v.", bi.Main.Version)

	_, _ = engines.NewDatabaseEngine()
}
