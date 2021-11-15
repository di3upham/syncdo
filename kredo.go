package syncdo

import "sync"

var krl = &sync.Mutex{}
var krm = make(map[string]int)

func KRedo(k string, f func()) int {
	var i int
	krl.Lock()
	if krm[k] > 0 {
		krm[k], i = krm[k]+1, krm[k]
		krl.Unlock()
		return i
	}
	krm[k], i = krm[k]+1, krm[k]
	krl.Unlock()
	defer func() { krl.Lock(); krm[k] = 0; krl.Unlock() }()
	f()
	return i
}
