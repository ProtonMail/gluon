package wait

import "sync"

type PanicHandler interface {
	HandlePanic()
}

type Group struct {
	wg           sync.WaitGroup
	PanicHandler PanicHandler
}

func (wg *Group) Go(f func()) {
	wg.wg.Add(1)

	go func() {
		defer wg.wg.Done()

		defer func() {
			if wg.PanicHandler != nil {
				wg.PanicHandler.HandlePanic()
			}
		}()

		f()
	}()
}

func (wg *Group) Wait() {
	wg.wg.Wait()
}
