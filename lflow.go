package syncdo

import "sync"

type Lflow struct {
	inc  chan interface{}
	errc chan error
	wg   *sync.WaitGroup
	earr [][]error
}

func NewLflow() *Lflow {
	return &Lflow{inc: make(chan interface{}), errc: make(chan error), wg: &sync.WaitGroup{}}
}

func (lf *Lflow) Run(num int, f func(i int, in interface{}) error) {
	if num <= 0 {
		return
	}
	from := len(lf.earr)
	lf.earr = append(lf.earr, make([][]error, num)...)
	lf.wg.Add(num)

	for i := from; i < num+from; i++ {
		go func(fi int) {
			defer func() { lf.wg.Done() }()
			var has bool
			var in interface{}
			var err error
			for {
				in, has = <-lf.inc
				if !has {
					break
				}
				err = f(fi, in)
				if err != nil {
					lf.earr[fi] = append(lf.earr[fi], err)
					select {
					case lf.errc <- err:
					default:
					}
					continue
				}
			}
		}(i)
	}
}

func (lf *Lflow) Push(in interface{}) bool {
	var err error
	select {
	case err = <-lf.errc:
	case lf.inc <- in:
	}
	return err == nil
}

func (lf *Lflow) Close() []error {
	close(lf.inc)
	lf.wg.Wait()
	close(lf.errc)

	arr := make([]error, 0)
	for _, a := range lf.earr {
		arr = append(arr, a...)
	}
	return arr
}
