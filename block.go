package syncdo

import "sync"

var ll2 = &sync.Mutex{}
var blm = make(map[string]*bmutex)
var bls = make([]*bmutex, 0)
var delc2 int

func BLock(b ...string) func() {
	ll2.Lock()
	var bl *bmutex
	for _, k := range b {
		if blm[k] != nil {
			bl = blm[k]
			break
		}
	}
	if bl == nil {
		bl = usebl()
	}

	var found bool
	for _, k := range b {
		found = false
		for _, k0 := range bl.set {
			if k0 == k {
				found = true
				break
			}
		}
		if !found {
			bl.set = append(bl.set, k)
		}
	}

	for _, k0 := range bl.set {
		blm[k0] = bl
	}

	bl.num++
	ll2.Unlock()
	bl.Lock()
	return func() {
		ll2.Lock()
		bl.num--
		if bl.num == 0 {
			for _, k0 := range bl.set {
				delete(blm, k0)
				delc2++
			}

			// go issue shrink map
			if delc2 > 8192 && len(blm) < 128 {
				m := make(map[string]*bmutex, len(blm))
				for k, v := range blm {
					m[k] = v
				}
				blm = m
				delc2 = 0
			}
		}
		ll2.Unlock()
		bl.Unlock()
	}
}

func usebl() *bmutex {
	for _, bl := range bls {
		if bl.num == 0 {
			return bl
		}
	}
	bl := &bmutex{Mutex: &sync.Mutex{}}
	bls = append(bls, bl)
	return bl
}

type bmutex struct {
	*sync.Mutex
	num int // number of waiters for the lock
	set []string
}
