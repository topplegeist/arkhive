package launcher

import (
	"sync"
)

type LauncherEngine struct {
}

func NewLauncherEngine() (instance *LauncherEngine, err error) {
	instance = &LauncherEngine{}
	return
}

func (launcherEngine *LauncherEngine) Initialize(waitGroup *sync.WaitGroup) {
}
