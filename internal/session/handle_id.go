package session

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func prepareServerResponse(info *internal.VersionInfo) response.Response {
	return response.ID(imap.NewIMAPIDFromVersionInfo(info))
}

func (s *Session) handleIDGet(ctx context.Context, tag string, ch chan response.Response) error {
	ch <- prepareServerResponse(s.version)
	ch <- response.Ok(tag).WithMessage("ID")

	return nil
}

func (s *Session) handleIDSet(ctx context.Context, tag string, cmd *proto.IDSet, ch chan response.Response) error {
	// Update session information
	s.imapID = imap.NewIMAPIDFromKeyMap(cmd.Keys)

	// Not logged in or no mailbox selected
	if s.state != nil {
		s.state.SetConnMetadataKeyValue(imap.IMAPIDConnMetadataKey, s.imapID)
	}

	ch <- prepareServerResponse(s.version)
	ch <- response.Ok(tag).WithMessage("ID")

	return nil
}
