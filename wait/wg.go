package wait

import "sync"

type Group struct {
	wg sync.WaitGroup
}

func (wg *Group) Go(f func()) {
	wg.wg.Add(1)

	go func() {
		defer wg.wg.Done()
		f()
	}()
}

func (wg *Group) Wait() {
	wg.wg.Wait()
}
