package syncdo

import (
	"fmt"
	"sync"
)

var ll = &sync.Mutex{}
var klm = make(map[string]*kmutex)
var kls = make([]*kmutex, 0)

func KLock(k string) func() {
	var kl *kmutex
	var has bool

	ll.Lock()
	kl, has = klm[k]
	if has {
		goto DO
	}
	kl = usekl()
	kl.key = &k
	klm[k] = kl

DO:
	kl.num++
	ll.Unlock()
	defer kl.Lock()
	return func() {
		ll.Lock()
		kl.rnum++
		if kl.rnum == kl.num {
			kl.key = nil
			kl.num = 0
			kl.rnum = 0
			delete(klm, k)
		}
		ll.Unlock()
		kl.Unlock()
	}
}

func Status() string {
	ll.Lock()
	m := len(klm)
	n := len(kls)
	ll.Unlock()
	return fmt.Sprintf("locking %d/%d", m, n)
}

func usekl() *kmutex {
	for _, kl := range kls {
		if kl.key == nil {
			return kl
		}
	}
	kl := &kmutex{Mutex: &sync.Mutex{}}
	kls = append(kls, kl)
	return kl
}

type kmutex struct {
	*sync.Mutex
	key  *string
	num  int
	rnum int
}
