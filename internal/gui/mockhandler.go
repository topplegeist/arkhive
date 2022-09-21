package gui

type MockHandler struct {
	IsStarted bool
}

func (mockHandler *MockHandler) NotifyStarted() {
	mockHandler.IsStarted = true
}
