package gluon

import "sync"

type wg struct {
	wg sync.WaitGroup
}

func (s *wg) Go(f func()) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		f()
	}()
}

func (s *wg) Wait() {
	s.wg.Wait()
}
