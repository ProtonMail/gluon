package gluon

import "sync"

type WaitGroup struct {
	wg sync.WaitGroup
}

func (wg *WaitGroup) Go(f func()) {
	wg.wg.Add(1)

	go func() {
		defer wg.wg.Done()
		f()
	}()
}

func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
}
