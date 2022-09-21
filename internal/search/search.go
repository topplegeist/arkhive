package search

import (
	"sync"
)

type SearchEngine struct {
}

func NewSearchEngine() (instance *SearchEngine, err error) {
	instance = &SearchEngine{}
	return
}

func (searchEngine *SearchEngine) Initialize(waitGroup *sync.WaitGroup) {}
