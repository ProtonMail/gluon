package state

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/contexts"
	"github.com/ProtonMail/gluon/internal/db"
	"github.com/ProtonMail/gluon/internal/db/ent"
	"github.com/ProtonMail/gluon/internal/parser/proto"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/bradenaw/juniper/xslices"
	"golang.org/x/text/encoding"
)

func (m *Mailbox) Search(ctx context.Context, keys []*proto.SearchKey, decoder *encoding.Decoder) ([]uint32, error) {
	snapMessages := m.snap.getAllMessages()

	messages, err := doSearch(ctx, m, snapMessages, keys, decoder)
	if err != nil {
		return nil, err
	}

	return xslices.Map(messages, func(msg snapMsgWithSeq) uint32 {
		if contexts.IsUID(ctx) {
			return uint32(msg.UID)
		}

		return uint32(msg.Seq)
	}), nil
}

func doSearch(ctx context.Context, m *Mailbox, candidates []snapMsgWithSeq, keys []*proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	for _, key := range keys {
		filtered, err := m.matchSearchKey(ctx, candidates, key, decoder)
		if err != nil {
			return nil, err
		}

		candidates = filtered
	}

	return candidates, nil
}

func (m *Mailbox) matchSearchKey(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	switch key.Keyword {
	case proto.SearchKeyword_SearchKWAll:
		return m.matchSearchKeyAll(ctx, candidates)

	case proto.SearchKeyword_SearchKWAnswered:
		return m.matchSearchKeyAnswered(ctx, candidates)

	case proto.SearchKeyword_SearchKWBcc:
		return m.matchSearchKeyBcc(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWBefore:
		return m.matchSearchKeyBefore(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWBody:
		return m.matchSearchKeyBody(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWCc:
		return m.matchSearchKeyCc(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWDeleted:
		return m.matchSearchKeyDeleted(ctx, candidates)

	case proto.SearchKeyword_SearchKWDraft:
		return m.matchSearchKeyDraft(ctx, candidates)

	case proto.SearchKeyword_SearchKWFlagged:
		return m.matchSearchKeyFlagged(ctx, candidates)

	case proto.SearchKeyword_SearchKWFrom:
		return m.matchSearchKeyFrom(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWHeader:
		return m.matchSearchKeyHeader(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWKeyword:
		return m.matchSearchKeyKeyword(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWLarger:
		return m.matchSearchKeyLarger(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWNew:
		return m.matchSearchKeyNew(ctx, candidates)

	case proto.SearchKeyword_SearchKWNot:
		return m.matchSearchKeyNot(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWOld:
		return m.matchSearchKeyOld(ctx, candidates)

	case proto.SearchKeyword_SearchKWOn:
		return m.matchSearchKeyOn(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWOr:
		return m.matchSearchKeyOr(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWRecent:
		return m.matchSearchKeyRecent(ctx, candidates)

	case proto.SearchKeyword_SearchKWSeen:
		return m.matchSearchKeySeen(ctx, candidates)

	case proto.SearchKeyword_SearchKWSentBefore:
		return m.matchSearchKeySentBefore(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWSentOn:
		return m.matchSearchKeySentOn(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWSentSince:
		return m.matchSearchKeySentSince(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWSince:
		return m.matchSearchKeySince(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWSmaller:
		return m.matchSearchKeySmaller(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWSubject:
		return m.matchSearchKeySubject(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWText:
		return m.matchSearchKeyText(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWTo:
		return m.matchSearchKeyTo(ctx, candidates, key, decoder)

	case proto.SearchKeyword_SearchKWUID:
		return m.matchSearchKeyUID(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWUnanswered:
		return m.matchSearchKeyUnanswered(ctx, candidates)

	case proto.SearchKeyword_SearchKWUndeleted:
		return m.matchSearchKeyUndeleted(ctx, candidates)

	case proto.SearchKeyword_SearchKWUndraft:
		return m.matchSearchKeyUndraft(ctx, candidates)

	case proto.SearchKeyword_SearchKWUnflagged:
		return m.matchSearchKeyUnflagged(ctx, candidates)

	case proto.SearchKeyword_SearchKWUnkeyword:
		return m.matchSearchKeyUnkeyword(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWUnseen:
		return m.matchSearchKeyUnseen(ctx, candidates)

	case proto.SearchKeyword_SearchKWSeqSet:
		return m.matchSearchKeySeqSet(ctx, candidates, key)

	case proto.SearchKeyword_SearchKWList:
		return m.matchSearchKeyList(ctx, candidates, key, decoder)

	default:
		return nil, fmt.Errorf("bad search keyword")
	}
}

func (m *Mailbox) matchSearchKeyAll(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return candidates, nil
}

func (m *Mailbox) matchSearchKeyAnswered(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagAnsweredLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyBcc(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "Bcc")
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return strings.Contains(strings.ToLower(value), strings.ToLower(string(b))), nil
	})
}

func (m *Mailbox) matchSearchKeyBefore(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	beforeDate, err := time.Parse("_2-Jan-2006", key.GetDate())
	if err != nil {
		return nil, err
	}

	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) ([]snapMsgWithSeq, error) {
		return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
			msg, err := db.GetMessage(ctx, client, message.ID.InternalID)
			if err != nil {
				return false, err
			}

			return msg.Date.Before(beforeDate), nil
		})
	})
}

func (m *Mailbox) matchSearchKeyBody(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		section := rfc822.Parse(literal)

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return bytes.Contains(bytes.ToLower(section.Body()), bytes.ToLower(b)), nil
	})
}

func (m *Mailbox) matchSearchKeyCc(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "Cc")
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return strings.Contains(strings.ToLower(value), strings.ToLower(string(b))), nil
	})
}

func (m *Mailbox) matchSearchKeyDeleted(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagDeletedLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyDraft(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagDraftLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyFlagged(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagFlaggedLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyFrom(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "From")
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return strings.Contains(strings.ToLower(value), strings.ToLower(string(b))), nil
	})
}

func (m *Mailbox) matchSearchKeyHeader(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, key.GetField())
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return strings.Contains(strings.ToLower(value), strings.ToLower(string(b))), nil
	})
}

func (m *Mailbox) matchSearchKeyKeyword(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	flagLowerCase := strings.ToLower(key.GetFlag())

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(flagLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyLarger(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) ([]snapMsgWithSeq, error) {
		return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
			msg, err := db.GetMessage(ctx, client, message.ID.InternalID)
			if err != nil {
				return false, err
			}

			return msg.Size > int(key.GetSize()), nil
		})
	})
}

func (m *Mailbox) matchSearchKeyNew(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagRecentLowerCase) && !message.flags.ContainsUnchecked(imap.FlagSeenLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyNot(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	left, err := m.matchSearchKey(ctx, candidates, key.GetLeftOp(), decoder)
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return xslices.IndexFunc(left, func(left snapMsgWithSeq) bool { return left.ID == message.ID }) < 0, nil
	})
}

func (m *Mailbox) matchSearchKeyOld(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(imap.FlagRecentLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyOn(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	onDate, err := time.Parse("_2-Jan-2006", key.GetDate())
	if err != nil {
		return nil, err
	}

	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) ([]snapMsgWithSeq, error) {
		return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
			msg, err := db.GetMessage(ctx, client, message.ID.InternalID)
			if err != nil {
				return false, err
			}

			return onDate.Truncate(24 * time.Hour).Equal(msg.Date.Truncate(24 * time.Hour)), nil
		})
	})
}

func (m *Mailbox) matchSearchKeyOr(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	left, err := m.matchSearchKey(ctx, candidates, key.GetLeftOp(), decoder)
	if err != nil {
		return nil, err
	}

	right, err := m.matchSearchKey(ctx, candidates, key.GetRightOp(), decoder)
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		leftIdx := xslices.IndexFunc(left, func(left snapMsgWithSeq) bool { return left.ID == message.ID })
		rightIdx := xslices.IndexFunc(right, func(right snapMsgWithSeq) bool { return right.ID == message.ID })

		return leftIdx >= 0 || rightIdx >= 0, nil
	})
}

func (m *Mailbox) matchSearchKeyRecent(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagRecentLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeySeen(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return message.flags.ContainsUnchecked(imap.FlagSeenLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeySentBefore(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	beforeDate, err := time.Parse("_2-Jan-2006", key.GetDate())
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "Date")
		if err != nil {
			return false, err
		}

		date, err := mail.ParseDate(value)
		if err != nil {
			return false, err
		}

		date = convertToDateWithoutTZ(date)

		return date.Before(beforeDate), nil
	})
}

func (m *Mailbox) matchSearchKeySentOn(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	onDate, err := time.Parse("_2-Jan-2006", key.GetDate())
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "Date")
		if err != nil {
			return false, err
		}

		date, err := mail.ParseDate(value)
		if err != nil {
			return false, err
		}

		// GODT-1598: Is this correct?
		date = convertToDateWithoutTZ(date)

		return date.Equal(onDate), nil
	})
}

func (m *Mailbox) matchSearchKeySentSince(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	sinceDate, err := time.Parse("_2-Jan-2006", key.GetDate())
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "Date")
		if err != nil {
			return false, err
		}

		date, err := mail.ParseDate(value)
		if err != nil {
			return false, err
		}

		date = convertToDateWithoutTZ(date)

		return date.Equal(sinceDate) || date.After(sinceDate), nil
	})
}

func (m *Mailbox) matchSearchKeySince(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	sinceDate, err := time.Parse("_2-Jan-2006", key.GetDate())
	if err != nil {
		return nil, err
	}

	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) ([]snapMsgWithSeq, error) {
		return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
			msg, err := db.GetMessage(ctx, client, message.ID.InternalID)
			if err != nil {
				return false, err
			}

			date := convertToDateWithoutTZ(msg.Date)

			return date.Equal(sinceDate) || date.After(sinceDate), nil
		})
	})
}

func (m *Mailbox) matchSearchKeySmaller(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	return db.ReadResult(ctx, m.state.db(), func(ctx context.Context, client *ent.Client) ([]snapMsgWithSeq, error) {
		return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
			msg, err := db.GetMessage(ctx, client, message.ID.InternalID)
			if err != nil {
				return false, err
			}

			return msg.Size < int(key.GetSize()), nil
		})
	})
}

func (m *Mailbox) matchSearchKeySubject(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "Subject")
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return strings.Contains(strings.ToLower(value), strings.ToLower(string(b))), nil
	})
}

func (m *Mailbox) matchSearchKeyText(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return bytes.Contains(bytes.ToLower(literal), bytes.ToLower(b)), nil
	})
}

func (m *Mailbox) matchSearchKeyTo(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		literal, err := m.state.getLiteral(message.ID.InternalID)
		if err != nil {
			return false, err
		}

		value, err := rfc822.GetHeaderValue(literal, "To")
		if err != nil {
			return false, err
		}

		b, err := decoder.Bytes(key.GetText())
		if err != nil {
			return false, err
		}

		return strings.Contains(strings.ToLower(value), strings.ToLower(string(b))), nil
	})
}

func (m *Mailbox) matchSearchKeyUID(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	left, err := m.snap.getMessagesInUIDRange(key.GetSequenceSet())
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return xslices.IndexFunc(left, func(left snapMsgWithSeq) bool { return left.ID == message.ID }) >= 0, nil
	})
}

func (m *Mailbox) matchSearchKeyUnanswered(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(imap.FlagAnsweredLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyUndeleted(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(imap.FlagDeletedLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyUndraft(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(imap.FlagDraftLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyUnflagged(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(imap.FlagFlaggedLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyUnkeyword(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	flagLowerCase := strings.ToLower(key.GetFlag())

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(flagLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeyUnseen(ctx context.Context, candidates []snapMsgWithSeq) ([]snapMsgWithSeq, error) {
	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return !message.flags.ContainsUnchecked(imap.FlagSeenLowerCase), nil
	})
}

func (m *Mailbox) matchSearchKeySeqSet(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey) ([]snapMsgWithSeq, error) {
	left, err := m.snap.getMessagesInSeqRange(key.GetSequenceSet())
	if err != nil {
		return nil, err
	}

	return filter(ctx, candidates, func(message snapMsgWithSeq) (bool, error) {
		return xslices.IndexFunc(left, func(left snapMsgWithSeq) bool { return left.ID == message.ID }) >= 0, nil
	})
}

func (m *Mailbox) matchSearchKeyList(ctx context.Context, candidates []snapMsgWithSeq, key *proto.SearchKey, decoder *encoding.Decoder) ([]snapMsgWithSeq, error) {
	return doSearch(ctx, m, candidates, key.GetChildren(), decoder)
}

func filter(ctx context.Context, candidates []snapMsgWithSeq, wantMessage func(seq snapMsgWithSeq) (bool, error)) ([]snapMsgWithSeq, error) {
	var res []snapMsgWithSeq

	for _, message := range candidates {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		default: // fallthrough
		}

		ok, err := wantMessage(message)
		if err != nil {
			return nil, err
		}

		if ok {
			res = append(res, message)
		}
	}

	return res, nil
}

func convertToDateWithoutTZ(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
