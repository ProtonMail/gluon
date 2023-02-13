package command

import (
	"fmt"
	"time"
)

type SearchKeyAll struct{}

func (s SearchKeyAll) String() string {
	return "ALL"
}

func (s SearchKeyAll) SanitizedString() string {
	return s.String()
}

type SearchKeyAnswered struct{}

func (s SearchKeyAnswered) String() string {
	return "ANSWERED"
}

func (s SearchKeyAnswered) SanitizedString() string {
	return s.String()
}

type SearchKeyBCC struct {
	Value string
}

func (s SearchKeyBCC) String() string {
	return fmt.Sprintf("BCC %v", s.Value)
}

func (s SearchKeyBCC) SanitizedString() string {
	return fmt.Sprintf("BCC %v", sanitizeString(s.Value))
}

type SearchKeyBefore struct {
	Value time.Time
}

func (s SearchKeyBefore) String() string {
	return fmt.Sprintf("BEFORE %v", s.Value)
}

func (s SearchKeyBefore) SanitizedString() string {
	return fmt.Sprintf("BEFORE <DATE>")
}

type SearchKeyBody struct {
	Value string
}

func (s SearchKeyBody) String() string {
	return fmt.Sprintf("BODY %v", s.Value)
}

func (s SearchKeyBody) SanitizedString() string {
	return fmt.Sprintf("BODY %v", sanitizeString(s.Value))
}

type SearchKeyCC struct {
	Value string
}

func (s SearchKeyCC) String() string {
	return fmt.Sprintf("CC %v", s.Value)
}

func (s SearchKeyCC) SanitizedString() string {
	return fmt.Sprintf("CC %v", sanitizeString(s.Value))
}

type SearchKeyDeleted struct{}

func (s SearchKeyDeleted) String() string {
	return "DELETED"
}

func (s SearchKeyDeleted) SanitizedString() string {
	return s.String()
}

type SearchKeyFlagged struct{}

func (s SearchKeyFlagged) String() string {
	return "Flagged"
}

func (s SearchKeyFlagged) SanitizedString() string {
	return s.String()
}

type SearchKeyFrom struct {
	Value string
}

func (s SearchKeyFrom) String() string {
	return fmt.Sprintf("FROM %v", s.Value)
}

func (s SearchKeyFrom) SanitizedString() string {
	return fmt.Sprintf("FROM %v", sanitizeString(s.Value))
}

type SearchKeyKeyword struct {
	Value string
}

func (s SearchKeyKeyword) String() string {
	return fmt.Sprintf("KEYWORD %v", s.Value)
}

func (s SearchKeyKeyword) SanitizedString() string {
	return fmt.Sprintf("KEYWORD %v", sanitizeString(s.Value))
}

type SearchKeyNew struct{}

func (s SearchKeyNew) String() string {
	return "NEW"
}

func (s SearchKeyNew) SanitizedString() string {
	return s.String()
}

type SearchKeyOld struct{}

func (s SearchKeyOld) String() string {
	return "NEW"
}

func (s SearchKeyOld) SanitizedString() string {
	return s.String()
}

type SearchKeyOn struct {
	Value time.Time
}

func (s SearchKeyOn) String() string {
	return fmt.Sprintf("On %v", s.Value)
}

func (s SearchKeyOn) SanitizedString() string {
	return fmt.Sprintf("ON <DATE>")
}

type SearchKeyRecent struct{}

func (s SearchKeyRecent) String() string {
	return "RECENT"
}

func (s SearchKeyRecent) SanitizedString() string {
	return s.String()
}

type SearchKeySeen struct{}

func (s SearchKeySeen) String() string {
	return "SEEN"
}

func (s SearchKeySeen) SanitizedString() string {
	return s.String()
}

type SearchKeySince struct {
	Value time.Time
}

func (s SearchKeySince) String() string {
	return fmt.Sprintf("SINCE %v", s.Value)
}

func (s SearchKeySince) SanitizedString() string {
	return fmt.Sprintf("SINCE <DATE>")
}

type SearchKeySubject struct {
	Value string
}

func (s SearchKeySubject) String() string {
	return fmt.Sprintf("SUBJECT %v", s.Value)
}

func (s SearchKeySubject) SanitizedString() string {
	return fmt.Sprintf("SUBJECT %v", sanitizeString(s.Value))
}

type SearchKeyText struct {
	Value string
}

func (s SearchKeyText) String() string {
	return fmt.Sprintf("TEXT %v", s.Value)
}

func (s SearchKeyText) SanitizedString() string {
	return fmt.Sprintf("TEXT %v", sanitizeString(s.Value))
}

type SearchKeyTo struct {
	Value string
}

func (s SearchKeyTo) String() string {
	return fmt.Sprintf("TO %v", s.Value)
}

func (s SearchKeyTo) SanitizedString() string {
	return fmt.Sprintf("TO %v", sanitizeString(s.Value))
}

type SearchKeyUnanswered struct{}

func (s SearchKeyUnanswered) String() string {
	return "UNANSWERED"
}

func (s SearchKeyUnanswered) SanitizedString() string {
	return s.String()
}

type SearchKeyUndeleted struct{}

func (s SearchKeyUndeleted) String() string {
	return "UNDELETED"
}

func (s SearchKeyUndeleted) SanitizedString() string {
	return s.String()
}

type SearchKeyUnflagged struct{}

func (s SearchKeyUnflagged) String() string {
	return "UNFLAGGED"
}

func (s SearchKeyUnflagged) SanitizedString() string {
	return s.String()
}

type SearchKeyUnkeyword struct {
	Value string
}

func (s SearchKeyUnkeyword) String() string {
	return fmt.Sprintf("UNKEYWORD %v", s.Value)
}

func (s SearchKeyUnkeyword) SanitizedString() string {
	return fmt.Sprintf("UNKEYWORD %v", sanitizeString(s.Value))
}

type SearchKeyUnseen struct{}

func (s SearchKeyUnseen) String() string {
	return "UNSEEN"
}

func (s SearchKeyUnseen) SanitizedString() string {
	return s.String()
}

type SearchKeyDraft struct{}

func (s SearchKeyDraft) String() string {
	return "DRAFT"
}

func (s SearchKeyDraft) SanitizedString() string {
	return s.String()
}

type SearchKeyHeader struct {
	Field string
	Value string
}

func (s SearchKeyHeader) String() string {
	return fmt.Sprintf("HEADER %v %v", s.Field, s.Value)
}

func (s SearchKeyHeader) SanitizedString() string {
	return fmt.Sprintf("HEADER %v %v", s.Field, sanitizeString(s.Value))
}

type SearchKeyLarger struct {
	Value int
}

func (s SearchKeyLarger) String() string {
	return fmt.Sprintf("LARGER %v", s.Value)
}

func (s SearchKeyLarger) SanitizedString() string {
	return s.String()
}

type SearchKeyNot struct {
	Key SearchKey
}

func (s SearchKeyNot) String() string {
	return fmt.Sprintf("NOT (%v)", s.Key.String())
}

func (s SearchKeyNot) SanitizedString() string {
	return fmt.Sprintf("NOT (%v)", s.Key.SanitizedString())
}

type SearchKeyOr struct {
	Key1 SearchKey
	Key2 SearchKey
}

func (s SearchKeyOr) String() string {
	return fmt.Sprintf("NOT ((%v) (%v))", s.Key1.String(), s.Key2.String())
}

func (s SearchKeyOr) SanitizedString() string {
	return fmt.Sprintf("NOT ((%v) (%v))", s.Key1.SanitizedString(), s.Key2.SanitizedString())
}

type SearchKeySentBefore struct {
	Value time.Time
}

func (s SearchKeySentBefore) String() string {
	return fmt.Sprintf("SENTBEFORE %v", s.Value)
}

func (s SearchKeySentBefore) SanitizedString() string {
	return fmt.Sprintf("SENTBEFORE <DATE>")
}

type SearchKeySentOn struct {
	Value time.Time
}

func (s SearchKeySentOn) String() string {
	return fmt.Sprintf("SENTON %v", s.Value)
}

func (s SearchKeySentOn) SanitizedString() string {
	return fmt.Sprintf("SENTON <DATE>")
}

type SearchKeySentSince struct {
	Value time.Time
}

func (s SearchKeySentSince) String() string {
	return fmt.Sprintf("SENTSINCE %v", s.Value)
}

func (s SearchKeySentSince) SanitizedString() string {
	return fmt.Sprintf("SENTSINCE <DATE>")
}

type SearchKeySmaller struct {
	Value int
}

func (s SearchKeySmaller) String() string {
	return fmt.Sprintf("SMALLER %v", s.Value)
}

func (s SearchKeySmaller) SanitizedString() string {
	return s.String()
}

type SearchKeyUID struct {
	Value uint32
}

func (s SearchKeyUID) String() string {
	return fmt.Sprintf("UID %v", s.Value)
}

func (s SearchKeyUID) SanitizedString() string {
	return s.String()
}

type SearchKeyUndraft struct{}

func (s SearchKeyUndraft) String() string {
	return "UNDRAFT"
}

func (s SearchKeyUndraft) SanitizedString() string {
	return s.String()
}
