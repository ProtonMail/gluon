package utils

import (
	"time"

	"github.com/ProtonMail/gluon/benchmarks/gluon_bench/flags"
	"github.com/emersion/go-imap/client"
)

// ScopedTimer tracks the duration between invocations to Start and Stop.
type ScopedTimer struct {
	start time.Time
	end   time.Time
}

func (s *ScopedTimer) Start() {
	s.start = time.Now()
}

func (s *ScopedTimer) Stop() {
	s.end = time.Now()
}

func (s *ScopedTimer) Elapsed() time.Duration {
	return s.end.Sub(s.start)
}

func FillBenchmarkSourceMailbox(cl *client.Client) error {
	if *flags.FillSourceMailboxFlag != 0 {
		if err := BuildMailbox(cl, *flags.MailboxFlag, int(*flags.FillSourceMailboxFlag)); err != nil {
			return err
		}
	}

	return nil
}
