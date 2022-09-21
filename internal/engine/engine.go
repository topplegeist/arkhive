package engine

import "sync"

type ApplicationEngine interface {
	Initialize(waitGroup *sync.WaitGroup)
}
