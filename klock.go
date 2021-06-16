package syncdo

func KLock(k string, n *int) {
	var kl *kmutex
	var has bool
	if *n > 0 {
		return
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
}

func KUnlock(k string, n *int) {
	ll.Lock()
	kl := klm[k]
	if *n == kl.num {
		kl.key = nil
		kl.num = 0
		delete(klm, k)
	}
	ll.Unlock()
	kl.Unlock()
}
