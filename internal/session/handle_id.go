package session

import (
	"context"

	"github.com/ProtonMail/gluon/events"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
)

func (s *Session) handleIDGet(ctx context.Context, tag string, ch chan response.Response) error {
	ch <- response.ID(imap.NewIMAPIDFromVersionInfo(s.version))

	ch <- response.Ok(tag).WithMessage("ID")

	return nil
}

func (s *Session) handleIDSet(ctx context.Context, tag string, cmd *proto.IDSet, ch chan response.Response) error {
	// Update session IMAP ID.
	s.imapID = imap.NewIMAPIDFromKeyMap(cmd.Keys)

	// If logged in and a mailbox has been selected, set the IMAP ID in the state's metadata.
	if s.state != nil {
		s.state.SetConnMetadataKeyValue(imap.IMAPIDConnMetadataKey, s.imapID)
	}

	ch <- response.ID(imap.NewIMAPIDFromVersionInfo(s.version))

	ch <- response.Ok(tag).WithMessage("ID")

	s.eventCh <- events.IMAPID{
		SessionID: s.sessionID,
		IMAPID:    s.imapID,
	}

	return nil
}
