package engine

import (
	"fmt"
	"sync"

	"arkhive.dev/launcher/internal/gui"
)

type Controller struct {
	engines                        []ApplicationEngine
	guiHandler                     gui.Handler
	coreThreadsInitializationGroup sync.WaitGroup
}

func NewController(engines []ApplicationEngine, guiHandler gui.Handler) (controller *Controller) {
	return &Controller{
		engines:    engines,
		guiHandler: guiHandler,
	}
}

func (controller *Controller) Initialize() {
	for engineIndex, engine := range controller.engines {
		if engine == nil {
			panic(fmt.Sprintf("Engine %d is nil", engineIndex))
		}
		controller.coreThreadsInitializationGroup.Add(1)
		go engine.Initialize(&controller.coreThreadsInitializationGroup)
	}

	controller.coreThreadsInitializationGroup.Wait()
	controller.guiHandler.NotifyStarted()
}
