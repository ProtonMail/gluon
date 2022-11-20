package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/profiling"
	"github.com/emersion/go-imap/utf7"
)

func (s *Session) handleSub(ctx context.Context, tag string, cmd *proto.Sub, ch chan response.Response) error {
	profiling.Start(ctx, profiling.CmdTypeSubscribe)
	defer profiling.Stop(ctx, profiling.CmdTypeSubscribe)

	nameUTF8, err := utf7.Encoding.NewDecoder().String(cmd.GetMailbox())
	if err != nil {
		return err
	}

	if err := s.state.Subscribe(ctx, nameUTF8); err != nil {
		return err
	}

	ch <- response.Ok(tag).WithMessage("SUB")

	return nil
}
