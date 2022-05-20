package syncdo

import (
	"sync/atomic"
	"time"
)

type Connin struct {
	i       int
	conn    interface{}
	isClose bool
	closeAt int64
	openAt  int64
	openIn  int64
	begin   int64
	end     int64
}

const MINT = 2 * time.Millisecond // chanable
const MAXT = 20 * time.Minute     // connable
const WAIT = 20 * time.Second     // waitable

type CPool struct {
	*Sand
	openConn  func() interface{}
	cap       int
	connins   []*Connin
	conchan   chan *Connin
	isNotfull bool
	waitc     int64
}

func NewCPool(cap int, openConn func() interface{}) *CPool {
	if cap < 2 {
		cap = 2
	}

	pool := &CPool{Sand: NewSand()}
	pool.openConn = openConn
	pool.cap = cap
	pool.connins = make([]*Connin, pool.cap)
	pool.conchan = make(chan *Connin, pool.cap)
	pool.isNotfull = true

	for i := 0; i < 2 && i < pool.cap; i++ {
		t0 := time.Now().UnixNano()
		conn := pool.openConn()
		if conn == nil {
			panic("open conn failure")
		}
		pool.connins[i] = &Connin{i: i, conn: conn, openAt: time.Now().UnixNano()}
		pool.connins[i].openIn = pool.connins[i].openAt - t0
		pool.conchan <- pool.connins[i]
	}

	return pool
}

func (pool *CPool) Use() (int, interface{}, func()) {
	atomic.AddInt64(&pool.waitc, 1)
	defer atomic.AddInt64(&pool.waitc, -1)
	if waitc := atomic.LoadInt64(&pool.waitc); waitc > int64(pool.cap) {
		var sumt int64
		var count int64
		for _, connin := range pool.connins {
			if connin != nil {
				sumt += connin.end - connin.begin
				count += 1
			}
		}
		if sumt*waitc > count*WAIT.Nanoseconds()*int64(pool.cap) {
			panic("overload")
		}
	}

	var connin *Connin

	select {
	case connin = <-pool.conchan:
	case <-time.After(MINT):
		if pool.isNotfull {
			go func() {
				si := pool.Stream(func() {})
				pool.Seep(si, func() {}, func() error {
					pool.appendConnin()
					time.Sleep(WAIT) // wait service
					return nil
				})
			}()
		}
	}

	if connin == nil {
		select {
		case connin = <-pool.conchan:
		case <-time.After(WAIT):
		}
	}

	if connin == nil {
		return 0, nil, func() {}
	}

	var begin, end int64
	begin = time.Now().UnixNano()
	return connin.i, connin.conn, func() {
		end = time.Now().UnixNano()
		if end-begin > WAIT.Nanoseconds() {
			// TODO warn
		}

		connin.begin, connin.end = begin, end

		select {
		case pool.conchan <- connin:
		}
	}
}

// callback
func (pool *CPool) Closed(index int) {
	pool.connins[index].isClose = true
	pool.connins[index].closeAt = time.Now().UnixNano()
	pool.isNotfull = true

	// still connable
	var connable bool
	t0 := time.Now().UnixNano()
	for _, connin := range pool.connins {
		if connin != nil && (!connin.isClose || t0-connin.closeAt < MAXT.Nanoseconds()) {
			connable = true
			break
		}
	}
	if !connable {
		panic("lost connection")
	}
}

func (pool *CPool) appendConnin() {
	for i := 0; i < pool.cap; i++ {
		if pool.connins[i] == nil {
			t0 := time.Now().UnixNano()
			if conn := pool.openConn(); conn != nil {
				pool.connins[i] = &Connin{i: i, conn: conn, openAt: time.Now().UnixNano()}
				pool.connins[i].openIn = pool.connins[i].openAt - t0
				pool.conchan <- pool.connins[i]
			}
			break
		}
		if pool.connins[i].isClose {
			t0 := time.Now().UnixNano()
			if conn := pool.openConn(); conn != nil {
				pool.connins[i].isClose = false
				pool.connins[i].conn = conn
				pool.connins[i].openAt = time.Now().UnixNano()
				pool.connins[i].openIn = pool.connins[i].openAt - t0
			}
			break
		}
	}

	pool.isNotfull = false
	for i := 0; i < pool.cap; i++ {
		if pool.connins[i] == nil || pool.connins[i].isClose {
			pool.isNotfull = true
			break
		}
	}
}
