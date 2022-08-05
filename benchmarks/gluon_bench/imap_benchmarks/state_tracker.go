package imap_benchmarks

import (
	"net"

	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
)

type stateTracker struct {
	MBoxes []string
}

func newStateTracker() *stateTracker {
	return &stateTracker{}
}

func (s *stateTracker) createRandomMBox(cl *client.Client) (string, error) {
	mbox := uuid.NewString()

	if err := cl.Create(mbox); err != nil {
		return "", err
	}

	s.MBoxes = append(s.MBoxes, mbox)

	return mbox, nil
}

func (s *stateTracker) createAndFillRandomMBox(cl *client.Client) (string, error) {
	mbox, err := s.createRandomMBox(cl)
	if err != nil {
		return "", err
	}

	if err := FillMailbox(cl, mbox); err != nil {
		return "", err
	}

	return mbox, nil
}

func (s *stateTracker) cleanup(cl *client.Client) error {
	for _, v := range s.MBoxes {
		if err := cl.Delete(v); err != nil {
			return err
		}
	}

	s.MBoxes = nil

	return nil
}

func (s *stateTracker) cleanupWithAddr(addr net.Addr) error {
	return WithClient(addr, func(c *client.Client) error {
		return s.cleanup(c)
	})
}
