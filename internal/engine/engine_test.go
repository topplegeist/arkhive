package engine_test

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type MockEngine struct {
	Index   uint
	Started bool
}

func (mockEngine *MockEngine) Initialize(waitGroup *sync.WaitGroup) {
	logrus.Infof("Mock engine %d started", mockEngine.Index)
	mockEngine.Started = true
	waitGroup.Done()
}
