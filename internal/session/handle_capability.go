package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleCapability(ctx context.Context, tag string, cmd *proto.Capability, ch chan response.Response) error {
	s.capsLock.Lock()
	defer s.capsLock.Unlock()

	ch <- response.Capability().WithCapabilities(s.caps...)

	ch <- response.Ok(tag).WithMessage("CAPABILITY")

	return nil
}
