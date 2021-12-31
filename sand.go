package syncdo

import "sync"

type Sand struct {
	*sync.Mutex
	sl    *sync.RWMutex
	index int
}

func NewSand() *Sand {
	return &Sand{Mutex: &sync.Mutex{}, sl: &sync.RWMutex{}}
}

func (s *Sand) Stream(f func()) int {
	s.Lock()
	defer s.Unlock()
	f()
	return s.index
}

func (s *Sand) Seep(si int, r func(), w func() error) error {
	s.Lock()
	if si < s.index {
		s.Unlock()
		s.sl.RLock()
		defer s.sl.RUnlock()
		return nil // not doing no error
	}
	s.index++
	r()
	s.sl.Lock()
	s.Unlock()
	defer s.sl.Unlock()
	return w()
}
