package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"github.com/ProtonMail/gluon/profiling"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

func (s *Session) handleSearch(ctx context.Context, tag string, cmd *proto.Search, mailbox *state.Mailbox, ch chan response.Response) (response.Response, error) {
	if contexts.IsUID(ctx) {
		profiling.Start(ctx, profiling.CmdTypeUIDSearch)
		defer profiling.Stop(ctx, profiling.CmdTypeUIDSearch)
	} else {
		profiling.Start(ctx, profiling.CmdTypeSearch)
		defer profiling.Stop(ctx, profiling.CmdTypeSearch)
	}

	var decoder *encoding.Decoder

	switch charset := cmd.GetOptionalCharset().(type) {
	case *proto.Search_Charset:
		encoding, err := ianaindex.IANA.Encoding(charset.Charset)
		if err != nil {
			return nil, response.No(tag).WithItems(response.ItemBadCharset())
		}

		decoder = encoding.NewDecoder()

	default:
		decoder = encoding.Nop.NewDecoder()
	}

	seq, err := mailbox.Search(ctx, cmd.GetKeys(), decoder)
	if err != nil {
		return nil, err
	}

	select {
	case ch <- response.Search(seq...):

	case <-ctx.Done():
		return nil, ctx.Err()
	}

	var items []response.Item

	if mailbox.ExpungeIssued() {
		items = append(items, response.ItemExpungeIssued())
	}

	return response.Ok(tag).
		WithItems(items...).
		WithMessage(okMessage(ctx)), nil
}
