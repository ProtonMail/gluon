package wait

import "sync"

type PanicHandler interface {
	HandlePanic()
}

type Group struct {
	wg           sync.WaitGroup
	panicHandler PanicHandler
}

func (wg *Group) SetPanicHandler(panicHandler PanicHandler) {
	wg.panicHandler = panicHandler
}

func (wg *Group) handlePanic() {
	if wg.panicHandler != nil {
		wg.panicHandler.HandlePanic()
	}
}

func (wg *Group) Go(f func()) {
	wg.wg.Add(1)

	go func() {
		defer wg.handlePanic()

		defer wg.wg.Done()
		f()
	}()
}

func (wg *Group) Wait() {
	wg.wg.Wait()
}
