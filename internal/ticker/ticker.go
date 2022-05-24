package ticker

import "time"

type Ticker struct {
	ticker *time.Ticker
	period time.Duration
	pollCh chan chan struct{}
	stopCh chan struct{}
}

func New(period time.Duration) *Ticker {
	return &Ticker{
		ticker: time.NewTicker(period),
		period: period,
		pollCh: make(chan chan struct{}),
		stopCh: make(chan struct{}),
	}
}

// Pause pauses the ticker.
func (ticker *Ticker) Pause() {
	ticker.ticker.Stop()
}

// Resume resumes the ticker.
func (ticker *Ticker) Resume() {
	ticker.ticker.Reset(ticker.period)
}

// Poll polls the ticker. It blocks until the tick has been executed.
func (ticker *Ticker) Poll() {
	doneCh := make(chan struct{})
	ticker.pollCh <- doneCh
	<-doneCh
}

// Stop stops the ticker.
func (ticker *Ticker) Stop() {
	close(ticker.stopCh)
}

// Tick calls the given callback at regular intervals or when the ticker is polled.
func (ticker *Ticker) Tick(fn func(time.Time)) {
	for {
		select {
		case tick := <-ticker.ticker.C:
			fn(tick)

		case doneCh := <-ticker.pollCh:
			fn(time.Now())
			close(doneCh)

		case <-ticker.stopCh:
			return
		}
	}
}
