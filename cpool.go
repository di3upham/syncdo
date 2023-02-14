package syncdo

import (
	"sync/atomic"
	"time"
)

type Conn interface {
	IsClosed() bool
}

type Connin struct {
	i     int
	conn  Conn
	begin int64
	end   int64
}

const MINT = 2 * time.Millisecond // chanable
const WAIT = 20 * time.Second     // waitable

type CPool struct {
	*Sand
	openConn  func() Conn
	cap       int
	connins   []*Connin
	conchan   chan *Connin
	isNotfull bool
	waitc     int64
}

func NewCPool(cap int, openConn func() Conn) *CPool {
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
		conn := pool.openConn()
		if conn == nil {
			panic("open conn failure")
		}
		pool.connins[i] = &Connin{i: i, conn: conn}
		pool.conchan <- pool.connins[i]
	}

	return pool
}

func (pool *CPool) Use() (int, Conn, func()) {
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
			go pool.appendConn()
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

	if connin.conn.IsClosed() {
		if conn := pool.openConn(); conn != nil {
			connin.conn = conn
		}
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

func (pool *CPool) appendConn() {
	si := pool.Stream(func() {})
	pool.Seep(si, func() {}, func() error {
		for i := 0; i < pool.cap; i++ {
			if pool.connins[i] == nil {
				if conn := pool.openConn(); conn != nil {
					pool.connins[i] = &Connin{i: i, conn: conn}
					pool.conchan <- pool.connins[i]
				}
				break
			}
		}
		pool.isNotfull = false
		for i := 0; i < pool.cap; i++ {
			if pool.connins[i] == nil {
				pool.isNotfull = true
				break
			}
		}
		return nil
	})
}
