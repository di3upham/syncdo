package syncdo

import "sync"

var ll = &sync.Mutex{}
var klm = make(map[string]*kmutex)
var kls = make([]*kmutex, 0)

func Line(k string, n *int, f func()) {
	var kl *kmutex
	var has bool
	if *n > 0 {
		goto F
	}

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
	*n = kl.num
	ll.Unlock()
	kl.Lock()
	defer func() {
		ll.Lock()
		if *n == kl.num {
			kl.key = nil
			kl.num = 0
			delete(klm, k)
		}
		ll.Unlock()
		kl.Unlock()
	}()
F:
	f()
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
	key *string
	num int
}
