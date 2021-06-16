package syncdo

import "fmt"

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
	n := new(int)
	*n = kl.num
	ll.Unlock()
	defer kl.Lock()
	return func() {
		ll.Lock()
		if *n == kl.num {
			kl.key = nil
			kl.num = 0
			delete(klm, k)
		}
		ll.Unlock()
		kl.Unlock()
	}
}

func Status() string {
	ll.Lock()
	runlen := len(klm)
	llen := len(kls)
	ll.Unlock()
	return fmt.Sprintf("locked-len=%d lock-total=%d", runlen, llen)
}
