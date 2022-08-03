package utils

import (
	"bufio"
	"os"
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
	if *flags.FillSourceMailbox != 0 {
		if err := BuildMailbox(cl, *flags.Mailbox, int(*flags.FillSourceMailbox)); err != nil {
			return err
		}
	}

	return nil
}

func ReadLinesFromFile(path string) ([]string, error) {
	readFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := make([]string, 0, 16)

	for fileScanner.Scan() {
		lines = append(lines, fileScanner.Text())
	}

	return lines, nil
}