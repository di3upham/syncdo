package syncdo

import (
	"testing"
	"time"
)

func TestCPoolUse(t *testing.T) {
	var exo, aco int
	p := NewCPool(
		20,
		func() Conn {
			return &testConn{}
		},
	)

	for i := 0; i < 100; i++ {
		go func(index int) {
			var conn *testConn
			_, conni, done := p.Use()
			conn, ok := conni.(*testConn)
			if !ok {
				t.Error("can not convert interface to testConn")
				return
			}
			if conn == nil {
				t.Error("conn is nil")
				return
			}
			defer done()
			conn.sumcall += 1
		}(i)
	}
	time.Sleep(110 * time.Millisecond)

	for _, connin := range p.connins {
		if connin == nil {
			continue
		}
		conn, ok := connin.conn.(*testConn)
		if !ok {
			t.Error("can not convert interface to testConn")
			return
		}
		if conn == nil {
			t.Error("conn is nil")
			return
		}
		aco += conn.sumcall
	}
	exo = 100
	if aco != exo {
		t.Errorf("want %d, actual %d", exo, aco)
	}
}

type testConn struct {
	sumcall  int
	isClosed bool
}

func (conn *testConn) IsClosed() bool {
	return conn.isClosed
}
