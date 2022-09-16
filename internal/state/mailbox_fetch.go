package state

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/ids"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
)

func (m *Mailbox) Fetch(ctx context.Context, seq *proto.SequenceSet, attributes []*proto.FetchAttribute, ch chan response.Response) error {
	msg, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return err
	}

	for _, msg := range msg {
		seq, items, err := m.fetchItems(ctx, msg, attributes)
		if err != nil {
			return err
		}

		select {
		case ch <- response.Fetch(seq).WithItems(items...):

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (m *Mailbox) fetchItems(ctx context.Context, msg *snapMsg, attributes []*proto.FetchAttribute) (imap.SeqID, []response.Item, error) {
	var (
		items []response.Item

		wantUID bool
		setSeen bool
	)

	message, err := db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (*ent.Message, error) {
		return db.GetMessage(ctx, client, msg.ID.InternalID)
	})
	if err != nil {
		return 0, nil, err
	}

	for _, attribute := range attributes {
		switch attribute := attribute.Attribute.(type) {
		case *proto.FetchAttribute_Keyword:
			item, err := m.fetchKeyword(msg, message, attribute.Keyword)
			if err != nil {
				return 0, nil, err
			}

			if attribute.Keyword == proto.FetchKeyword_FetchKWUID {
				wantUID = true
			}

			if attribute.Keyword == proto.FetchKeyword_FetchKWRFC822 || attribute.Keyword == proto.FetchKeyword_FetchKWRFC822Text {
				setSeen = true
			}

			items = append(items, item)

		case *proto.FetchAttribute_Body:
			literal, err := m.state.getLiteral(msg.ID.InternalID)
			if err != nil {
				return 0, nil, err
			}

			item, err := m.fetchBody(attribute.Body, literal)
			if err != nil {
				return 0, nil, err
			}

			items = append(items, item)

			if !attribute.Body.Peek {
				setSeen = true
			}
		}
	}

	if contexts.IsUID(ctx) && !wantUID {
		items = append(items, response.ItemUID(msg.UID))
	}

	if setSeen {
		newFlags, err := db.WriteResult(ctx, m.state.db(), func(ctx context.Context, tx *ent.Tx) (map[imap.InternalMessageID]imap.FlagSet, error) {
			return m.state.actionAddMessageFlags(ctx, tx, []ids.MessageIDPair{msg.ID}, imap.NewFlagSet(imap.FlagSeen))
		})
		if err != nil {
			return 0, nil, err
		}

		newMessageFlags := newFlags[msg.ID.InternalID]

		if !msg.flags.Equals(newMessageFlags) {
			if err := m.snap.setMessageFlags(msg.ID.InternalID, newMessageFlags); err != nil {
				return 0, nil, err
			}

			items = append(items, response.ItemFlags(newMessageFlags))
		}
	}

	return msg.Seq, items, nil
}

func (m *Mailbox) fetchKeyword(msg *snapMsg, message *ent.Message, keyword proto.FetchKeyword) (response.Item, error) {
	switch keyword {
	case proto.FetchKeyword_FetchKWEnvelope:
		return response.ItemEnvelope(message.Envelope), nil

	case proto.FetchKeyword_FetchKWFlags:
		return response.ItemFlags(msg.flags), nil

	case proto.FetchKeyword_FetchKWInternalDate:
		return response.ItemInternalDate(message.Date), nil

	case proto.FetchKeyword_FetchKWRFC822:
		return m.fetchRFC822(msg.ID.InternalID)

	case proto.FetchKeyword_FetchKWRFC822Header:
		return m.fetchRFC822Header(msg.ID.InternalID)

	case proto.FetchKeyword_FetchKWRFC822Size:
		return response.ItemRFC822Size(message.Size), nil

	case proto.FetchKeyword_FetchKWRFC822Text:
		return m.fetchRFC822Text(msg.ID.InternalID)

	case proto.FetchKeyword_FetchKWBody:
		return response.ItemBody(message.Body), nil

	case proto.FetchKeyword_FetchKWBodyStructure:
		return response.ItemBodyStructure(message.BodyStructure), nil

	case proto.FetchKeyword_FetchKWUID:
		return response.ItemUID(msg.UID), nil

	default:
		return nil, fmt.Errorf("bad fetch keyword")
	}
}

func (m *Mailbox) fetchRFC822(messageID imap.InternalMessageID) (response.Item, error) {
	literal, err := m.state.getLiteral(messageID)
	if err != nil {
		return nil, err
	}

	return response.ItemRFC822Literal(literal), nil
}

func (m *Mailbox) fetchRFC822Header(messageID imap.InternalMessageID) (response.Item, error) {
	literal, err := m.state.getLiteral(messageID)
	if err != nil {
		return nil, err
	}

	section := rfc822.Parse(literal)

	return response.ItemRFC822Header(section.Header()), nil
}

func (m *Mailbox) fetchRFC822Text(messageID imap.InternalMessageID) (response.Item, error) {
	literal, err := m.state.getLiteral(messageID)
	if err != nil {
		return nil, err
	}

	section := rfc822.Parse(literal)

	return response.ItemRFC822Text(section.Body()), nil
}

func (m *Mailbox) fetchBody(body *proto.FetchBody, literal []byte) (response.Item, error) {
	b, section, err := m.fetchBodyLiteral(body, literal)
	if err != nil {
		return nil, err
	}

	item := response.ItemBodyLiteral(section, b)

	switch partial := body.GetOptionalPartial().(type) {
	case *proto.FetchBody_Partial:
		item.WithPartial(int(partial.Partial.GetBegin()), int(partial.Partial.GetCount()))
	}

	return item, nil
}

func (m *Mailbox) fetchBodyLiteral(body *proto.FetchBody, literal []byte) ([]byte, string, error) {
	switch section := body.OptionalSection.(type) {
	case *proto.FetchBody_Section:
		b, err := m.fetchBodySection(section.Section, literal)
		if err != nil {
			return nil, "", err
		}

		renderedSection, err := renderSection(section.Section)
		if err != nil {
			return nil, "", err
		}

		return b, renderedSection, nil

	default:
		return literal, "", nil
	}
}

func (m *Mailbox) fetchBodySection(section *proto.BodySection, literal []byte) ([]byte, error) {
	root := rfc822.Parse(literal)

	if parts := intParts(section.Parts); len(parts) > 0 {
		p, err := root.Part(parts...)
		if err != nil {
			return nil, err
		}

		root = p
	}

	if root == nil {
		return nil, fmt.Errorf("invalid section part")
	}

	switch keyword := section.OptionalKeyword.(type) {
	case *proto.BodySection_Keyword:
		// HEADER and TEXT keywords should handle embedded message/rfc822 parts!
		if keyword.Keyword != proto.SectionKeyword_MIME {
			contentType, _, err := root.ContentType()
			if err != nil {
				return nil, err
			}

			if rfc822.MIMEType(contentType) == rfc822.MessageRFC822 {
				root = rfc822.Parse(root.Body())
			}
		}

		switch keyword.Keyword {
		case proto.SectionKeyword_Header:
			return root.Header(), nil

		case proto.SectionKeyword_HeaderFields:
			header, err := root.ParseHeader()
			if err != nil {
				return nil, err
			}

			return header.Fields(section.Fields), nil

		case proto.SectionKeyword_HeaderFieldsNot:
			header, err := root.ParseHeader()
			if err != nil {
				return nil, err
			}

			return header.FieldsNot(section.Fields), nil

		case proto.SectionKeyword_Text:
			return root.Body(), nil

		case proto.SectionKeyword_MIME:
			return root.Header(), nil

		default:
			return nil, fmt.Errorf("bad section keyword")
		}

	default:
		return root.Body(), nil
	}
}

func renderSection(section *proto.BodySection) (string, error) {
	var res []string

	if len(section.Parts) > 0 {
		res = append(res, renderParts(intParts(section.Parts)))
	}

	switch keyword := section.GetOptionalKeyword().(type) {
	case *proto.BodySection_Keyword:
		switch keyword.Keyword {
		case proto.SectionKeyword_Header:
			res = append(res, "HEADER")

		case proto.SectionKeyword_HeaderFields:
			res = append(res, fmt.Sprintf("HEADER.FIELDS (%v)", strings.Join(section.GetFields(), " ")))

		case proto.SectionKeyword_HeaderFieldsNot:
			res = append(res, fmt.Sprintf("HEADER.FIELDS.NOT (%v)", strings.Join(section.GetFields(), " ")))

		case proto.SectionKeyword_Text:
			res = append(res, "TEXT")

		case proto.SectionKeyword_MIME:
			res = append(res, "MIME")

		default:
			return "", fmt.Errorf("bad body section keyword")
		}
	}

	return strings.ToUpper(strings.Join(res, ".")), nil
}

func renderParts(sectionParts []int) string {
	return strings.Join(xslices.Map(sectionParts, func(part int) string { return strconv.Itoa(part) }), ".")
}

func intParts(parts []int32) []int {
	return xslices.Map(parts, func(part int32) int { return int(part) })
}
