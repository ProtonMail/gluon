package state

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/imap/command"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/response"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/parallel"
	"github.com/bradenaw/juniper/xslices"
)

var totalActiveFetchRequest int32

func (m *Mailbox) Fetch(ctx context.Context, cmd *command.Fetch, ch chan response.Response) error {
	snapMessages, err := m.snap.getMessagesInRange(ctx, cmd.SeqSet)
	if err != nil {
		return err
	}

	operations := make([]func(snapMsgWithSeq, *ent.Message, []byte) (response.Item, error), 0, len(cmd.Attributes))

	var (
		needsLiteral bool
		wantUID      bool
		setSeen      bool
	)

	for _, attribute := range cmd.Attributes {
		switch attribute := attribute.(type) {
		case *command.FetchAttributeAll:
			// Macro equivalent to: (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE).
			operations = append(operations, fetchFlags, fetchInternalDate, fetchRFC822Size, fetchEnvelope)
		case *command.FetchAttributeFast:
			// Macro equivalent to: (FLAGS INTERNALDATE RFC822.SIZE).
			operations = append(operations, fetchFlags, fetchInternalDate, fetchRFC822Size)
		case *command.FetchAttributeFull:
			// Macro equivalent to: (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY).
			operations = append(operations, fetchFlags, fetchInternalDate, fetchRFC822Size, fetchEnvelope, fetchBody)
		case *command.FetchAttributeUID:
			wantUID = true

			operations = append(operations, fetchUID)
		case *command.FetchAttributeRFC822:
			setSeen = true
			needsLiteral = true

			operations = append(operations, fetchRFC822)
		case *command.FetchAttributeRFC822Text:
			setSeen = true
			needsLiteral = true

			operations = append(operations, fetchRFC822Text)
		case *command.FetchAttributeRFC822Header:
			needsLiteral = true

			operations = append(operations, fetchRFC822Header)
		case *command.FetchAttributeRFC822Size:
			operations = append(operations, fetchRFC822Size)
		case *command.FetchAttributeFlags:
			operations = append(operations, fetchFlags)
		case *command.FetchAttributeEnvelope:
			operations = append(operations, fetchEnvelope)
		case *command.FetchAttributeInternalDate:
			operations = append(operations, fetchInternalDate)
		case *command.FetchAttributeBody:
			operations = append(operations, fetchBody)
		case *command.FetchAttributeBodyStructure:
			operations = append(operations, fetchBodyStructure)
		case *command.FetchAttributeBodySection:
			needsLiteral = true

			if !attribute.Peek {
				setSeen = true
			}

			op := func(_ snapMsgWithSeq, _ *ent.Message, literal []byte) (response.Item, error) {
				return fetchAttributeBodySection(attribute, literal)
			}

			operations = append(operations, op)
		}
	}

	const minCountForParallelism = 4

	var parallelism int

	activeFetchRequests := atomic.AddInt32(&totalActiveFetchRequest, 1)
	defer atomic.AddInt32(&totalActiveFetchRequest, -1)

	// Only run in parallel if we have to fetch more than minCountForParallelism messages or if we have more than one
	// message and we need to access the literal.
	if !contexts.IsParallelismDisabledCtx(ctx) && (len(snapMessages) > minCountForParallelism || (len(snapMessages) > 1 && needsLiteral)) {
		// If multiple fetch request are happening in parallel, reduce the number of goroutines in proportion to that
		// to avoid overloading the user's machine.
		parallelism = runtime.NumCPU() / int(activeFetchRequests)

		// make sure that if division hits 0, we run single threaded rather than use MAXGOPROCS
		if parallelism < 1 {
			parallelism = 1
		}
	} else {
		parallelism = 1
	}

	if err := parallel.DoContext(ctx, parallelism, len(snapMessages), func(ctx context.Context, i int) error {
		msg := snapMessages[i]
		message, err := db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) (*ent.Message, error) {
			return db.GetMessage(ctx, client, msg.ID.InternalID)
		})
		if err != nil {
			return err
		}

		var literal []byte

		if needsLiteral {
			l, err := m.state.getLiteral(ctx, msg.ID)
			if err != nil {
				return err
			}

			literal = l
		}

		items := make([]response.Item, 0, len(operations))

		for _, op := range operations {
			item, err := op(msg, message, literal)
			if err != nil {
				return err
			}

			items = append(items, item)
		}

		if contexts.IsUID(ctx) && !wantUID {
			items = append(items, response.ItemUID(msg.UID))
		}

		if setSeen {
			if !msg.flags.ContainsUnchecked(imap.FlagSeenLowerCase) {
				msg.flags.AddToSelf(imap.FlagSeen)

				items = append(items, response.ItemFlags(msg.flags))

			}
		} else {
			// remove message from the list to avoid being processed for seen flag changes later.
			snapMessages[i].snapMsg = nil
		}

		ch <- response.Fetch(msg.Seq).WithItems(items...)

		return nil
	}); err != nil {
		return err
	}

	msgsToBeMarkedSeen := xslices.Filter(snapMessages, func(s snapMsgWithSeq) bool {
		return s.snapMsg != nil
	})

	if len(msgsToBeMarkedSeen) != 0 {
		if err := m.state.db().Write(ctx, func(ctx context.Context, tx *ent.Tx) error {
			return m.state.actionAddMessageFlags(ctx, tx, msgsToBeMarkedSeen, imap.NewFlagSet(imap.FlagSeen))
		}); err != nil {
			return err
		}
	}

	return nil
}

func fetchEnvelope(_ snapMsgWithSeq, message *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemEnvelope(message.Envelope), nil
}

func fetchFlags(msg snapMsgWithSeq, message *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemFlags(msg.flags), nil
}

func fetchInternalDate(_ snapMsgWithSeq, message *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemInternalDate(message.Date), nil
}

func fetchRFC822(_ snapMsgWithSeq, _ *ent.Message, literal []byte) (response.Item, error) {
	return response.ItemRFC822Literal(literal), nil
}

func fetchRFC822Header(_ snapMsgWithSeq, _ *ent.Message, literal []byte) (response.Item, error) {
	section := rfc822.Parse(literal)

	return response.ItemRFC822Header(section.Header()), nil
}

func fetchRFC822Size(_ snapMsgWithSeq, message *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemRFC822Size(message.Size), nil
}

func fetchRFC822Text(_ snapMsgWithSeq, _ *ent.Message, literal []byte) (response.Item, error) {
	section := rfc822.Parse(literal)

	return response.ItemRFC822Text(section.Body()), nil
}

func fetchBody(_ snapMsgWithSeq, message *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemBody(message.Body), nil
}

func fetchBodyStructure(_ snapMsgWithSeq, message *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemBodyStructure(message.BodyStructure), nil
}

func fetchUID(msg snapMsgWithSeq, _ *ent.Message, _ []byte) (response.Item, error) {
	return response.ItemUID(msg.UID), nil
}

func fetchAttributeBodySection(attribute *command.FetchAttributeBodySection, literal []byte) (response.Item, error) {
	b, section, err := fetchBodyLiteral(attribute.Section, literal)
	if err != nil {
		return nil, err
	}

	item := response.ItemBodyLiteral(section, b)

	if attribute.Partial != nil {
		item.WithPartial(int(attribute.Partial.Offset), int(attribute.Partial.Count))
	}

	return item, nil
}

func fetchBodyLiteral(section command.BodySection, literal []byte) ([]byte, string, error) {
	if section == nil {
		return literal, "", nil
	}

	b, err := fetchBodySection(section, literal)
	if err != nil {
		return nil, "", err
	}

	renderedSection, err := renderSection(section)
	if err != nil {
		return nil, "", err
	}

	return b, renderedSection, nil
}

func fetchBodySection(section command.BodySection, literal []byte) ([]byte, error) {
	root := rfc822.Parse(literal)

	switch v := section.(type) {
	case *command.BodySectionPart:
		if len(v.Part) > 0 {
			p, err := root.Part(v.Part...)
			if err != nil {
				return nil, err
			}

			root = p
			section = v.Section
		}
	default:
	}

	if root == nil {
		return nil, fmt.Errorf("invalid section part")
	}

	// HEADER and TEXT keywords should handle embedded message/rfc822 parts!
	handleEmbeddedParts := func(root *rfc822.Section) (*rfc822.Section, error) {
		contentType, _, err := root.ContentType()
		if err != nil {
			return nil, err
		}

		if rfc822.MIMEType(contentType) == rfc822.MessageRFC822 {
			root = rfc822.Parse(root.Body())
		}

		return root, nil
	}

	if section == nil {
		return root.Body(), nil
	}

	switch section := section.(type) {
	case *command.BodySectionMIME:
		return root.Header(), nil
	case *command.BodySectionHeader:
		r, err := handleEmbeddedParts(root)
		if err != nil {
			return nil, err
		}

		return r.Header(), nil
	case *command.BodySectionText:
		r, err := handleEmbeddedParts(root)
		if err != nil {
			return nil, err
		}

		return r.Body(), nil
	case *command.BodySectionHeaderFields:
		r, err := handleEmbeddedParts(root)
		if err != nil {
			return nil, err
		}

		header, err := r.ParseHeader()
		if err != nil {
			return nil, err
		}

		if section.Negate {
			return header.FieldsNot(section.Fields), nil
		}

		return header.Fields(section.Fields), nil
	default:
		return nil, fmt.Errorf("unknown section")
	}
}

func renderSection(section command.BodySection) (string, error) {
	var res []string

	switch v := section.(type) {
	case *command.BodySectionPart:
		res = append(res, renderParts(v.Part))
		section = v.Section
	default:
	}

	if section != nil {
		switch section := section.(type) {
		case *command.BodySectionMIME:
			res = append(res, "MIME")
		case *command.BodySectionHeader:
			res = append(res, "HEADER")
		case *command.BodySectionText:
			res = append(res, "TEXT")
		case *command.BodySectionHeaderFields:
			if section.Negate {
				res = append(res, fmt.Sprintf("HEADER.FIELDS.NOT (%v)", strings.Join(section.Fields, " ")))
			} else {
				res = append(res, fmt.Sprintf("HEADER.FIELDS (%v)", strings.Join(section.Fields, " ")))
			}
		default:
			return "", fmt.Errorf("bad body section keyword")
		}
	}

	return strings.ToUpper(strings.Join(res, ".")), nil
}

func renderParts(sectionParts []int) string {
	return strings.Join(xslices.Map(sectionParts, func(part int) string { return strconv.Itoa(part) }), ".")
}
