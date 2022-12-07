package syncdo

import (
	"sync"
	"time"
)

// debug lock, useful thing for newbie
// the example below

/*
l := NewDLock()
l.Lock(1000, func() { fmt.Println("TOO-LONGGGGGGGGGGGGG") })
defer l.Unlock()
// function body
*/

type DLock struct {
	l     *sync.Mutex
	index int
}

func NewDLock() *DLock {
	return &DLock{l: &sync.Mutex{}}
}

func (dl *DLock) Lock(ms int, f func()) {
	dl.l.Lock()
	go func(i int) {
		for {
			time.Sleep(time.Duration(ms) * time.Millisecond)
			if dl.index > i {
				break
			}
			f()
		}
	}(dl.index)
}

func (dl *DLock) Unlock() {
	dl.index++
	dl.l.Unlock()
}
