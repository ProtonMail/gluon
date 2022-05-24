package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleLogout(ctx context.Context, tag string, cmd *proto.Logout) error {
	if err := response.Bye().Send(s); err != nil {
		return err
	}

	if err := response.Ok(tag).Send(s); err != nil {
		return err
	}

	return nil
}
