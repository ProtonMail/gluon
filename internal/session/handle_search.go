package session

import (
	"context"

	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/internal/state"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

func (s *Session) handleSearch(ctx context.Context, tag string, cmd *proto.Search, mailbox *state.Mailbox, ch chan response.Response) error {
	var decoder *encoding.Decoder

	switch charset := cmd.GetOptionalCharset().(type) {
	case *proto.Search_Charset:
		encoding, err := ianaindex.IANA.Encoding(charset.Charset)
		if err != nil {
			return response.No(tag).WithItems(response.ItemBadCharset())
		}

		decoder = encoding.NewDecoder()

	default:
		decoder = encoding.Nop.NewDecoder()
	}

	seq, err := mailbox.Search(ctx, cmd.GetKeys(), decoder)
	if err != nil {
		return err
	}

	ch <- response.Search(seq...)

	var items []response.Item

	if mailbox.ExpungeIssued() {
		items = append(items, response.ItemExpungeIssued())
	}

	ch <- response.Ok(tag).
		WithItems(items...).
		WithMessage(okMessage(ctx))

	return nil
}
