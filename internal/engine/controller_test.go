package engine_test

import (
	"fmt"
	"testing"

	"arkhive.dev/launcher/internal/engine"
	"github.com/stretchr/testify/assert"
)

type MockHandler struct {
	IsStarted bool
}

func (mockHandler *MockHandler) NotifyStarted() {
	mockHandler.IsStarted = true
}

func TestInitializeNoEngines(t *testing.T) {
	engines := make([]engine.ApplicationEngine, 0)
	handler := MockHandler{}
	controller := engine.NewController(engines, &handler)
	controller.Initialize()
	assert.True(t, handler.IsStarted, "The mock GUI not notifies the start")
}

func TestInitialize(t *testing.T) {
	const enginesCount = 5
	engines := make([]engine.ApplicationEngine, enginesCount)

	for engineIndex := uint(0); engineIndex < enginesCount; engineIndex++ {
		engines[engineIndex] = &MockEngine{Index: engineIndex}
	}

	handler := MockHandler{}

	controller := engine.NewController(engines, &handler)
	controller.Initialize()

	for engineIndex := 0; engineIndex < enginesCount; engineIndex++ {
		assert.True(t, engines[engineIndex].(*MockEngine).Started, fmt.Sprintf("The mock engine %d not started", engineIndex))
	}
	assert.True(t, handler.IsStarted, "The mock GUI not notifies the start")
}
