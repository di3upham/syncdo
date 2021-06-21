package syncdo

import (
	"sync"
	"time"
)

type CPool struct {
	*sync.Mutex
	conchan  chan *Connin
	connins  []*Connin
	makeconn func() interface{}
}

func NewCPool(max int, makeconn func() interface{}) *CPool {
	if max < 2 {
		max = 2
	}
	conchan := make(chan *Connin, max)
	connins := make([]*Connin, max)
	for i := 0; i < 2 && i < max; i++ {
		conn := makeconn()
		if conn == nil {
			continue
		}
		connins[i] = &Connin{index: i, conn: conn}
		conchan <- connins[i]
	}
	return &CPool{Mutex: &sync.Mutex{}, conchan: conchan, connins: connins, makeconn: makeconn}
}

func (pool *CPool) Use(tout time.Duration) (interface{}, func()) {
	var connin *Connin

	select {
	case connin = <-pool.conchan:
	case <-time.After(MINT):
		connin = pool.newconnin()
	}

	if connin == nil {
		if tout > MINT {
			select {
			case connin = <-pool.conchan:
			case <-time.After(tout):
			}
		}
		if tout <= 0 {
			select {
			case connin = <-pool.conchan:
			}
		}
	}

	if connin == nil {
		return nil, func() {}
	}

	connin.begin = time.Now().UnixNano()

	return connin.conn, func() {
		connin.end = time.Now().UnixNano()

		select {
		case pool.conchan <- connin:
		}
	}
}

func (pool *CPool) newconnin() *Connin {
	pool.Lock()
	defer pool.Unlock()

	connlen := len(pool.connins)
	for i := 0; i < connlen; i++ {
		if pool.connins[i] == nil {
			if connin := pool.makeconn(); connin != nil {
				pool.connins[i] = &Connin{index: i, conn: connin}
				return pool.connins[i]
			}
			break
		}
	}

	return nil
}

type Connin struct {
	index int
	conn  interface{}
	begin int64
	end   int64
}

const MINT = 1 * time.Millisecond

// TODO crud connin
// TODO pool status
