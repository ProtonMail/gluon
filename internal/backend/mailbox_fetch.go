package backend

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/backend/ent"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/stream"
	"github.com/bradenaw/juniper/xslices"
)

var maxFetchConcurrency = runtime.NumCPU()

func (m *Mailbox) Fetch(ctx context.Context, seq *proto.SequenceSet, attributes []*proto.FetchAttribute, ch chan response.Response) error {
	msg, err := m.snap.getMessagesInRange(ctx, seq)
	if err != nil {
		return err
	}

	type res struct {
		seq   int
		items []response.Item
	}

	resCh := parallel.MapStream(
		ctx,
		newUIDStream(m.snap, m.mbox, msg),
		maxFetchConcurrency,
		maxFetchConcurrency,
		func(ctx context.Context, msg *ent.UID) (*res, error) {
			seq, items, err := m.fetchItems(ctx, msg, attributes)
			if err != nil {
				return nil, err
			}

			return &res{seq: seq, items: items}, nil
		},
	)
	defer resCh.Close()

	for {
		res, err := resCh.Next(ctx)
		if errors.Is(err, stream.End) {
			return nil
		} else if err != nil {
			return err
		}

		ch <- response.Fetch(res.seq).WithItems(res.items...)
	}
}

func (m *Mailbox) fetchItems(ctx context.Context, entMsg *ent.UID, attributes []*proto.FetchAttribute) (int, []response.Item, error) {
	var (
		items []response.Item

		wantUID bool
		setSeen bool
	)

	msg, err := m.snap.getMessage(entMsg.Edges.Message.MessageID)
	if err != nil {
		return 0, nil, err
	}

	for _, attribute := range attributes {
		switch attribute := attribute.Attribute.(type) {
		case *proto.FetchAttribute_Keyword:
			item, err := m.fetchKeyword(msg, entMsg, attribute.Keyword)
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
			literal, err := m.state.getLiteral(msg.ID)
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

	if isUID(ctx) && !wantUID {
		items = append(items, response.ItemUID(msg.UID))
	}

	if setSeen {
		newFlags, err := m.state.actionAddMessageFlags(ctx, m.tx, []string{msg.ID}, imap.NewFlagSet(imap.FlagSeen))
		if err != nil {
			return 0, nil, err
		}

		if !msg.flags.Equals(newFlags[msg.ID]) {
			items = append(items, response.ItemFlags(newFlags[msg.ID]))
		}
	}

	return msg.Seq, items, nil
}

func (m *Mailbox) fetchKeyword(msg *snapMsg, res *ent.UID, keyword proto.FetchKeyword) (response.Item, error) {
	switch keyword {
	case proto.FetchKeyword_FetchKWEnvelope:
		return response.ItemEnvelope(res.Edges.Message.Envelope), nil

	case proto.FetchKeyword_FetchKWFlags:
		return response.ItemFlags(msg.flags), nil

	case proto.FetchKeyword_FetchKWInternalDate:
		return response.ItemInternalDate(res.Edges.Message.Date), nil

	case proto.FetchKeyword_FetchKWRFC822:
		return m.fetchRFC822(msg.ID)

	case proto.FetchKeyword_FetchKWRFC822Header:
		return m.fetchRFC822Header(msg.ID)

	case proto.FetchKeyword_FetchKWRFC822Size:
		return response.ItemRFC822Size(res.Edges.Message.Size), nil

	case proto.FetchKeyword_FetchKWRFC822Text:
		return m.fetchRFC822Text(msg.ID)

	case proto.FetchKeyword_FetchKWBody:
		return response.ItemBody(res.Edges.Message.Body), nil

	case proto.FetchKeyword_FetchKWBodyStructure:
		return response.ItemBodyStructure(res.Edges.Message.BodyStructure), nil

	case proto.FetchKeyword_FetchKWUID:
		return response.ItemUID(res.UID), nil

	default:
		panic("bad keyword")
	}
}

func (m *Mailbox) fetchRFC822(messageID string) (response.Item, error) {
	literal, err := m.state.getLiteral(messageID)
	if err != nil {
		return nil, err
	}

	return response.ItemRFC822Literal(literal), nil
}

func (m *Mailbox) fetchRFC822Header(messageID string) (response.Item, error) {
	literal, err := m.state.getLiteral(messageID)
	if err != nil {
		return nil, err
	}

	section, err := rfc822.Parse(literal)
	if err != nil {
		return nil, err
	}

	return response.ItemRFC822Header(section.Header()), nil
}

func (m *Mailbox) fetchRFC822Text(messageID string) (response.Item, error) {
	literal, err := m.state.getLiteral(messageID)
	if err != nil {
		return nil, err
	}

	section, err := rfc822.Parse(literal)
	if err != nil {
		return nil, err
	}

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

		return b, renderSection(section.Section), nil

	default:
		return literal, "", nil
	}
}

func (m *Mailbox) fetchBodySection(section *proto.BodySection, literal []byte) ([]byte, error) {
	root, err := rfc822.Parse(literal)
	if err != nil {
		return nil, err
	}

	if parts := intParts(section.Parts); len(parts) > 0 {
		root = root.Part(parts...)
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
				if root, err = rfc822.Parse(root.Body()); err != nil {
					return nil, err
				}
			}
		}

		switch keyword.Keyword {
		case proto.SectionKeyword_Header:
			return root.Header(), nil

		case proto.SectionKeyword_HeaderFields:
			return root.ParseHeader().Fields(section.Fields), nil

		case proto.SectionKeyword_HeaderFieldsNot:
			return root.ParseHeader().FieldsNot(section.Fields), nil

		case proto.SectionKeyword_Text:
			return root.Body(), nil

		case proto.SectionKeyword_MIME:
			return root.Header(), nil

		default:
			panic("bad keyword")
		}

	default:
		return root.Body(), nil
	}
}

func renderSection(section *proto.BodySection) string {
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
			panic("bad keyword")
		}
	}

	return strings.ToUpper(strings.Join(res, "."))
}

func renderParts(sectionParts []int) string {
	return strings.Join(xslices.Map(sectionParts, func(part int) string { return strconv.Itoa(part) }), ".")
}

func intParts(parts []int32) []int {
	return xslices.Map(parts, func(part int32) int { return int(part) })
}
