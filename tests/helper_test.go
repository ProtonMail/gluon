package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/internal/backend"
	"github.com/bradenaw/juniper/iterator"
	"github.com/bradenaw/juniper/xslices"
	goimap "github.com/emersion/go-imap"
	uidplus "github.com/emersion/go-imap-uidplus"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testHelper struct {
	s   *testConnection
	tag string
}

func (h *testHelper) expect(res ...string) {
	res[len(res)-1] = fmt.Sprintf(`%v %v`, h.tag, res[len(res)-1])

	h.s.Sxe(res...)
}

func (s *testConnection) doAppend(mailboxName, literal string, flags ...string) *testHelper {
	tag := uuid.NewString()

	s.Cf(`%v APPEND %v (%v) {%v}`, tag, mailboxName, strings.Join(flags, " "), len(literal))
	s.Sx(`\+.*`)
	s.C(literal)

	return &testHelper{s: s, tag: tag}
}

func (s *testConnection) doAppendFromFile(mailboxName, path string, flags ...string) *testHelper {
	literal, err := os.ReadFile(path)
	require.NoError(s.tb, err)

	return s.doAppend(mailboxName, string(literal), flags...)
}

func doAppendWithClient(client *client.Client, mailboxName string, literal string, time time.Time, flags ...string) error {
	return client.Append(mailboxName, flags, time, strings.NewReader(literal))
}

func doAppendWithClientFromFile(t testing.TB, client *client.Client, mailboxName string, filePath string, time time.Time, flags ...string) error {
	literal, err := os.ReadFile(filePath)
	require.NoError(t, err)

	return doAppendWithClient(client, mailboxName, string(literal), time, flags...)
}

func doAppendWithClientPlus(client *uidplus.Client, mailboxName string, literal string, flags ...string) (uint32, uint32, error) {
	return client.Append(mailboxName, flags, time.Now(), strings.NewReader(literal))
}

func doAppendWithClientPlusFromFile(t testing.TB, client *uidplus.Client, mailboxName string, filePath string, flags ...string) (uint32, uint32, error) {
	literal, err := os.ReadFile(filePath)
	require.NoError(t, err)

	return doAppendWithClientPlus(client, mailboxName, string(literal), flags...)
}

func listMailboxesClient(t testing.TB, client *client.Client, reference string, expression string) []*goimap.MailboxInfo {
	mailboxesChannel := make(chan *goimap.MailboxInfo)
	done := make(chan error, 1)

	go func() {
		done <- client.List(reference, expression, mailboxesChannel)
		require.NoError(t, <-done)
	}()

	return iterator.Collect(iterator.Chan(mailboxesChannel))
}

func checkMailboxesMatchNamesAndAttributes(t *testing.T, client *client.Client, reference string, expression string, expectedNames []string, expectedAttributes []string) {
	mailboxes := listMailboxesClient(t, client, "", "*")

	var actualMailboxNames []string

	for _, mailbox := range mailboxes {
		actualMailboxNames = append(actualMailboxNames, mailbox.Name)
		require.ElementsMatch(t, mailbox.Attributes, expectedAttributes)
	}

	require.ElementsMatch(t, actualMailboxNames, expectedNames)
}

func getMailboxNamesClient(t testing.TB, client *client.Client, reference string, expression string) []string {
	mailboxes := listMailboxesClient(t, client, reference, expression)

	return xslices.Map(mailboxes, func(mbox *goimap.MailboxInfo) string {
		return mbox.Name
	})
}

func matchMailboxNamesClient(t testing.TB, client *client.Client, reference string, expression string, expectedNames []string) {
	mailboxes := getMailboxNamesClient(t, client, reference, expression)
	require.ElementsMatch(t, expectedNames, mailboxes)
}

func storeWithRetrievalClient(t testing.TB, client *client.Client, seqset *goimap.SeqSet, item goimap.StoreItem, value interface{}) []*goimap.Message {
	ch := make(chan *goimap.Message)

	go func() {
		require.NoError(t, client.Store(seqset, item, value, ch))
	}()

	return iterator.Collect(iterator.Chan(ch))
}

func uidStoreWithRetrievalClient(t testing.TB, client *client.Client, seqset *goimap.SeqSet, item goimap.StoreItem, value interface{}) []*goimap.Message {
	ch := make(chan *goimap.Message)

	go func() {
		require.NoError(t, client.UidStore(seqset, item, value, ch))
	}()

	return iterator.Collect(iterator.Chan(ch))
}

type macroFetchCommand struct {
	SeqSet *goimap.SeqSet
	Item   goimap.FetchItem
}

func (cmd *macroFetchCommand) Command() *goimap.Command {
	return &goimap.Command{
		Name:      "FETCH",
		Arguments: []interface{}{cmd.SeqSet, goimap.RawString(cmd.Item)},
	}
}

func fetchMessagesClient(t testing.TB, client *client.Client, seqset *goimap.SeqSet, items []goimap.FetchItem) []*goimap.Message {
	ch := make(chan *goimap.Message)

	go func() {
		require.NoError(t, client.Fetch(seqset, items, ch))
	}()

	return iterator.Collect(iterator.Chan(ch))
}

func uidFetchMessagesClient(t testing.TB, client *client.Client, seqset *goimap.SeqSet, items []goimap.FetchItem) []*goimap.Message {
	ch := make(chan *goimap.Message)

	go func() {
		require.NoError(t, client.UidFetch(seqset, items, ch))
	}()

	return iterator.Collect(iterator.Chan(ch))
}

func expungeClient(t testing.TB, client *client.Client) []uint32 {
	expungeCh := make(chan uint32)

	go func() {
		require.NoError(t, client.Expunge(expungeCh))
	}()

	return iterator.Collect(iterator.Chan(expungeCh))
}

func uidExpungeClient(t testing.TB, client *uidplus.Client, sequenceSet *goimap.SeqSet) []uint32 {
	expungeCh := make(chan uint32)

	go func() {
		require.NoError(t, client.UidExpunge(sequenceSet, expungeCh))
	}()

	return iterator.Collect(iterator.Chan(expungeCh))
}

func createSeqSet(sequence string) *goimap.SeqSet {
	sequenceSet, err := goimap.ParseSeqSet(sequence)
	if err != nil {
		panic(err)
	}

	return sequenceSet
}

// Helper to validate go-imap-client's message Envelope.
type envelopeValidator struct {
	validateDateTime  func(testing.TB, time.Time)
	validateSubject   func(testing.TB, string)
	validateFrom      func(testing.TB, []*goimap.Address)
	validateSender    func(testing.TB, []*goimap.Address)
	validateTo        func(testing.TB, []*goimap.Address)
	validateReplyTo   func(testing.TB, []*goimap.Address)
	validateCc        func(testing.TB, []*goimap.Address)
	validateBcc       func(testing.TB, []*goimap.Address)
	validateMessageId func(testing.TB, string)
	validateInReplyTo func(testing.TB, string)
}

// newEmptyEnvelopValidator returns a validator that ensures all fields in the envelope are empty and have not
// been populated by a response.
// Note: MessageId is not checked by default.
func newEmptyEnvelopValidator() *envelopeValidator {
	return &envelopeValidator{
		validateDateTime: func(t testing.TB, t2 time.Time) {
			require.Zero(t, t2)
		},
		validateSubject: func(t testing.TB, s string) {
			require.Empty(t, s)
		},
		validateFrom: func(t testing.TB, addresses []*goimap.Address) {
			require.Empty(t, addresses)
		},
		validateTo: func(t testing.TB, addresses []*goimap.Address) {
			require.Empty(t, addresses)
		},
		validateSender: func(t testing.TB, addresses []*goimap.Address) {
			require.Empty(t, addresses)
		},
		validateReplyTo: func(t testing.TB, addresses []*goimap.Address) {
			require.Empty(t, addresses)
		},
		validateCc: func(t testing.TB, addresses []*goimap.Address) {
			require.Empty(t, addresses)
		},
		validateBcc: func(t testing.TB, addresses []*goimap.Address) {
			require.Empty(t, addresses)
		},
		validateMessageId: nil,
		validateInReplyTo: func(t testing.TB, s string) {
			require.Empty(t, s)
		},
	}
}

func (validator *envelopeValidator) check(t testing.TB, envelope *goimap.Envelope) {
	validator.validateDateTime(t, envelope.Date)
	validator.validateSubject(t, envelope.Subject)
	validator.validateFrom(t, envelope.From)
	validator.validateTo(t, envelope.To)
	validator.validateCc(t, envelope.Cc)
	validator.validateBcc(t, envelope.Bcc)
	validator.validateSender(t, envelope.Sender)

	if validator.validateMessageId != nil {
		validator.validateMessageId(t, envelope.MessageId)
	}

	validator.validateInReplyTo(t, envelope.InReplyTo)
}

// Helper to validate go-imap-client's message. When the fields `validateEnvelope`, `validateBodyStructure`,
// `validateInternalDate`, `validateBody` are nil, it implies that those fields were never set in the Message.
type messageValidator struct {
	validateSeqNum        func(testing.TB, uint32)
	validateUid           func(testing.TB, uint32)
	validateEnvelope      *envelopeValidator
	validateBodyStructure *bodyStructureValidator
	validateFlags         func(testing.TB, []string)
	validateInternalDate  func(testing.TB, time.Time)
	validateSize          func(testing.TB, uint32)
	validateBody          []func(testing.TB, *goimap.Message)
}

// newEmptyIMAPMessageValidator returns a validator that performs default checks for when the message has no data that
// was parsed as part of a response.
// Note: The fields `validateInternalDate`, `validateEnvelope`, `validateBodyStructure` and `validateBody` are always
// set to nil.
func newEmptyIMAPMessageValidator() *messageValidator {
	return &messageValidator{
		validateSeqNum: func(t testing.TB, u uint32) {
			require.Greater(t, u, uint32(0))
		},
		validateUid: func(t testing.TB, u uint32) {
			require.Zero(t, u)
		},
		validateEnvelope:      nil,
		validateBodyStructure: nil,
		validateFlags: func(t testing.TB, flags []string) {
			require.Empty(t, flags)
		},
		validateInternalDate: nil,
		validateSize: func(t testing.TB, u uint32) {
			require.Zero(t, u)
		},
		validateBody: nil,
	}
}

func (validator *messageValidator) check(t testing.TB, message *goimap.Message) {
	validator.validateSeqNum(t, message.SeqNum)
	validator.validateUid(t, message.Uid)

	if validator.validateEnvelope != nil {
		require.NotNil(t, message.Envelope)
		validator.validateEnvelope.check(t, message.Envelope)
	} else {
		require.Nil(t, message.Envelope)
	}

	validator.validateFlags(t, message.Flags)

	if validator.validateInternalDate != nil {
		validator.validateInternalDate(t, message.InternalDate)
	}

	if validator.validateBodyStructure != nil {
		require.NotNil(t, message.BodyStructure)
		validator.validateBodyStructure.check(t, message.BodyStructure)
	} else {
		require.Nil(t, message.BodyStructure)
	}

	validator.validateSize(t, message.Size)

	for _, fn := range validator.validateBody {
		fn(t, message)
	}
}

type fetchCommand struct {
	t      testing.TB
	client *client.Client
	items  []goimap.FetchItem
}

func newFetchCommand(t testing.TB, client *client.Client) *fetchCommand {
	return &fetchCommand{
		t:      t,
		client: client,
		items:  nil,
	}
}

func (fc *fetchCommand) withItems(items ...goimap.FetchItem) *fetchCommand {
	fc.items = xslices.Join(items)
	return fc
}

func (fc *fetchCommand) fetch(seqSet string) *fetchResult {
	return newFetchResult(fc.t, fetchMessagesClient(fc.t, fc.client, createSeqSet(seqSet), fc.items))
}

func (fc *fetchCommand) fetchFailure(seqSet string) {
	// we have to create the channel because the client code doesn't properly handle nil channels
	require.Error(fc.t, fc.client.Fetch(createSeqSet(seqSet), fc.items, make(chan *goimap.Message)))
}

func (fc *fetchCommand) fetchUid(seqSet string) *fetchResult {
	return newFetchResult(fc.t, uidFetchMessagesClient(fc.t, fc.client, createSeqSet(seqSet), fc.items))
}

type fetchResult struct {
	t          testing.TB
	messages   []*goimap.Message
	validators []*messageValidator
}

func newFetchResult(t testing.TB, messages []*goimap.Message) *fetchResult {
	return &fetchResult{
		t:        t,
		messages: messages,
		validators: xslices.Map(messages, func(_ *goimap.Message) *messageValidator {
			return nil
		}),
	}
}

type validatorBuilder messageValidator

func (fr *fetchResult) forSeqNum(number uint32, builderCallback func(*validatorBuilder)) *fetchResult {
	for index, message := range fr.messages {
		if message.SeqNum == number {
			validator := newEmptyIMAPMessageValidator()
			builderCallback((*validatorBuilder)(validator))
			fr.validators[index] = validator

			return fr
		}
	}

	panic("Could not locate message with given sequence number")
}

func (fr *fetchResult) forUid(id uint32, builderCallback func(*validatorBuilder)) *fetchResult {
	for index, message := range fr.messages {
		if message.Uid == id {
			validator := newEmptyIMAPMessageValidator()
			// No need to validate UID, we already verified this at this point
			validator.validateUid = func(_ testing.TB, _ uint32) {}
			builderCallback((*validatorBuilder)(validator))
			fr.validators[index] = validator

			return fr
		}
	}

	panic("Could not locate message with given sequence number")
}

func (fr *fetchResult) check() {
	for index, validator := range fr.validators {
		if validator != nil {
			validator.check(fr.t, fr.messages[index])
		}
	}
}

func (fr *fetchResult) checkAndRequireMessageCount(messageCount int) {
	require.Equal(fr.t, len(fr.messages), messageCount)
	fr.check()
}

func (vb *validatorBuilder) wantSequenceId(sequenceId uint32) *validatorBuilder {
	vb.validateSeqNum = func(t testing.TB, u uint32) {
		require.Equal(t, u, sequenceId)
	}

	return vb
}

func (vb *validatorBuilder) wantUID(uid uint32) *validatorBuilder {
	vb.validateUid = func(t testing.TB, u uint32) {
		require.Equal(t, u, uid)
	}

	return vb
}

func (vb *validatorBuilder) wantFlags(flags ...string) *validatorBuilder {
	vb.validateFlags = func(t testing.TB, f []string) {
		require.ElementsMatch(t, xslices.Map(flags, strings.ToLower), xslices.Map(f, strings.ToLower))
	}

	return vb
}

func (vb *validatorBuilder) ignoreFlags() *validatorBuilder {
	vb.validateFlags = func(_ testing.TB, _ []string) {}
	return vb
}

func (vb *validatorBuilder) wantInternalDate(dateTime string) *validatorBuilder {
	vb.validateInternalDate = func(t testing.TB, t2 time.Time) {
		expectedTime, err := time.Parse(goimap.DateTimeLayout, dateTime)
		require.NoError(t, err)
		require.Equal(t, expectedTime, t2)
	}

	return vb
}

func (vb *validatorBuilder) wantSize(size uint32) *validatorBuilder {
	vb.validateSize = func(t testing.TB, u uint32) {
		require.Equal(t, u, size)
	}

	return vb
}

func (vb *validatorBuilder) wantBodyStructure(fn func(builder *bodyStructureValidatorBuilder)) *validatorBuilder {
	validator := newEmptyBodyStructureValidator()
	fn((*bodyStructureValidatorBuilder)(validator))
	vb.validateBodyStructure = validator

	return vb
}

// There is a bug in the goimage library where the call to `section.resp()` always sets Peek member variable
// of the section to false, this causes issues with Peek requests, as the section will not be found. This
// implementation is equal to GetBody(), except we don't call `section.resp()`.
func getBodySection(message *goimap.Message, section *goimap.BodySectionName) goimap.Literal {
	for s, body := range message.Body {
		if section.Equal(s) {
			if body == nil {
				// Server can return nil, we need to treat as empty string per RFC 3501
				body = bytes.NewReader(nil)
			}

			return body
		}
	}

	return nil
}

func skipGLUONHeader(message string) string {
	if keyIndex := strings.Index(message, backend.InternalIDKey); keyIndex != -1 {
		message = message[0:keyIndex] + message[keyIndex+backend.InternalIDHeaderLengthWithNewLine:]
	}

	return message
}

func (vb *validatorBuilder) wantSection(sectionStr goimap.FetchItem, lines ...string) *validatorBuilder {
	section, err := goimap.ParseBodySectionName(sectionStr)
	if err != nil {
		panic("Failed to parse section string")
	}

	vb.validateBody = append(vb.validateBody, func(t testing.TB, message *goimap.Message) {
		literal := getBodySection(message, section)
		require.NotNil(t, literal)
		bytes, err := io.ReadAll(literal)
		require.NoError(t, err)
		require.Equal(t, string(bytes), strings.Join(lines, "\r\n"))
	})

	return vb
}

func (vb *validatorBuilder) wantSectionEmpty(sectionStr goimap.FetchItem) *validatorBuilder {
	section, err := goimap.ParseBodySectionName(sectionStr)
	if err != nil {
		panic("Failed to parse section string")
	}

	vb.validateBody = append(vb.validateBody, func(t testing.TB, message *goimap.Message) {
		literal := getBodySection(message, section)
		require.Nil(t, literal)
	})

	return vb
}

func (vb *validatorBuilder) wantSectionNotEmpty(sectionStr goimap.FetchItem) *validatorBuilder {
	section, err := goimap.ParseBodySectionName(sectionStr)
	if err != nil {
		panic("Failed to parse section string")
	}

	vb.validateBody = append(vb.validateBody, func(t testing.TB, message *goimap.Message) {
		literal := getBodySection(message, section)
		require.NotEmpty(t, literal)
	})

	return vb
}

func (vb *validatorBuilder) wantSectionAndSkipGLUONHeader(sectionStr goimap.FetchItem, expected ...string) *validatorBuilder {
	section, err := goimap.ParseBodySectionName(sectionStr)
	if err != nil {
		panic("Failed to parse section string")
	}

	vb.validateBody = append(vb.validateBody, func(t testing.TB, message *goimap.Message) {
		literal := getBodySection(message, section)
		require.NotNil(t, literal)
		bytes, err := io.ReadAll(literal)
		require.NoError(t, err)
		require.Equal(t, skipGLUONHeader(string(bytes)), strings.Join(expected, "\r\n"))
	})

	return vb
}

func (vb *validatorBuilder) wantSectionBytes(sectionStr goimap.FetchItem, fn func(testing.TB, []byte)) *validatorBuilder {
	section, err := goimap.ParseBodySectionName(sectionStr)
	if err != nil {
		panic("Failed to parse section string")
	}

	vb.validateBody = append(vb.validateBody, func(t testing.TB, message *goimap.Message) {
		literal := getBodySection(message, section)
		require.NotNil(t, literal, "Failed to get literal for section: %v", sectionStr)
		bytes, err := io.ReadAll(literal)
		require.NoError(t, err)
		fn(t, bytes)
	})

	return vb
}

func (vb *validatorBuilder) wantSectionString(sectionStr goimap.FetchItem, fn func(testing.TB, string)) *validatorBuilder {
	return vb.wantSectionBytes(sectionStr, func(t testing.TB, bytes []byte) {
		fn(t, string(bytes))
	})
}

type envelopeValidatorBuilder struct {
	validator *envelopeValidator
}

func (v *envelopeValidatorBuilder) wantDateTime(dateTime string) *envelopeValidatorBuilder {
	v.validator.validateDateTime = func(t testing.TB, t2 time.Time) {
		expectedTime, err := time.Parse(goimap.DateTimeLayout, dateTime)
		require.NoError(t, err)
		require.Equal(t, expectedTime, t2)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantSubject(subject string) *envelopeValidatorBuilder {
	v.validator.validateSubject = func(t testing.TB, s string) {
		require.Equal(t, s, subject)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantFrom(imapAddresses ...string) *envelopeValidatorBuilder {
	v.validator.validateFrom = func(t testing.TB, addresses []*goimap.Address) {
		messageAddresses := xslices.Map(addresses, (*goimap.Address).Address)
		require.ElementsMatch(t, imapAddresses, messageAddresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantAddressTypeFrom(imapAddresses ...*goimap.Address) *envelopeValidatorBuilder {
	v.validator.validateFrom = func(t testing.TB, addresses []*goimap.Address) {
		require.ElementsMatch(t, imapAddresses, addresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantAddressTypeSender(imapAddresses ...*goimap.Address) *envelopeValidatorBuilder {
	v.validator.validateSender = func(t testing.TB, addresses []*goimap.Address) {
		require.ElementsMatch(t, imapAddresses, addresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantSender(imapAddresses ...string) *envelopeValidatorBuilder {
	v.validator.validateSender = func(t testing.TB, addresses []*goimap.Address) {
		messageAddresses := xslices.Map(addresses, (*goimap.Address).Address)
		require.ElementsMatch(t, imapAddresses, messageAddresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantAddressTypeReplyTo(imapAddresses ...*goimap.Address) *envelopeValidatorBuilder {
	v.validator.validateReplyTo = func(t testing.TB, addresses []*goimap.Address) {
		require.ElementsMatch(t, imapAddresses, addresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantReplyTo(imapAddresses ...string) *envelopeValidatorBuilder {
	v.validator.validateReplyTo = func(t testing.TB, addresses []*goimap.Address) {
		messageAddresses := xslices.Map(addresses, (*goimap.Address).Address)
		require.ElementsMatch(t, imapAddresses, messageAddresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantTo(imapAddresses ...string) *envelopeValidatorBuilder {
	v.validator.validateTo = func(t testing.TB, addresses []*goimap.Address) {
		messageAddresses := xslices.Map(addresses, (*goimap.Address).Address)
		require.ElementsMatch(t, imapAddresses, messageAddresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantAddressTypeTo(imapAddresses ...*goimap.Address) *envelopeValidatorBuilder {
	v.validator.validateTo = func(t testing.TB, addresses []*goimap.Address) {
		require.ElementsMatch(t, imapAddresses, addresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantBcc(imapAddresses ...string) *envelopeValidatorBuilder {
	v.validator.validateBcc = func(t testing.TB, addresses []*goimap.Address) {
		messageAddresses := xslices.Map(addresses, (*goimap.Address).Address)
		require.ElementsMatch(t, imapAddresses, messageAddresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantAddressTypeBcc(imapAddresses ...*goimap.Address) *envelopeValidatorBuilder {
	v.validator.validateBcc = func(t testing.TB, addresses []*goimap.Address) {
		require.ElementsMatch(t, imapAddresses, addresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantCc(imapAddresses ...string) *envelopeValidatorBuilder {
	v.validator.validateCc = func(t testing.TB, addresses []*goimap.Address) {
		messageAddresses := xslices.Map(addresses, (*goimap.Address).Address)
		require.ElementsMatch(t, imapAddresses, messageAddresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantAddressTypeCc(imapAddresses ...*goimap.Address) *envelopeValidatorBuilder {
	v.validator.validateCc = func(t testing.TB, addresses []*goimap.Address) {
		require.ElementsMatch(t, imapAddresses, addresses)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantMessageId(id string) *envelopeValidatorBuilder {
	v.validator.validateMessageId = func(t testing.TB, msgId string) {
		require.Equal(t, msgId, id)
	}

	return v
}

func (v *envelopeValidatorBuilder) wantInReplyTo(value string) *envelopeValidatorBuilder {
	v.validator.validateInReplyTo = func(t testing.TB, s string) {
		require.Equal(t, value, s)
	}

	return v
}

func (vb *validatorBuilder) wantEnvelope(fn func(*envelopeValidatorBuilder)) *validatorBuilder {
	builder := &envelopeValidatorBuilder{
		validator: newEmptyEnvelopValidator(),
	}
	fn(builder)
	vb.validateEnvelope = builder.validator

	return vb
}

type bodyStructureValidator struct {
	validateMIMEType          func(testing.TB, string)
	validateMimeSubType       func(testing.TB, string)
	validateParams            func(testing.TB, map[string]string)
	validateID                func(testing.TB, string)
	validateDescription       func(testing.TB, string)
	validateEncoding          func(testing.TB, string)
	validateSize              func(testing.TB, uint32)
	validateParts             []*bodyStructureValidator
	validateRFCEnvelope       *envelopeValidator
	validateRFCBodyStructure  *bodyStructureValidator
	validateLines             func(testing.TB, uint32)
	validateExtended          func(testing.TB, bool)
	validateDisposition       func(testing.TB, string)
	validateDispositionParams func(testing.TB, map[string]string)
	validateLanguage          func(testing.TB, []string)
	validateLocation          func(testing.TB, []string)
	validateMD5               func(testing.TB, string)
}

func (bsv *bodyStructureValidator) check(tb testing.TB, structure *goimap.BodyStructure) {
	bsv.validateMIMEType(tb, structure.MIMEType)
	bsv.validateMimeSubType(tb, structure.MIMESubType)
	bsv.validateParams(tb, structure.Params)
	bsv.validateID(tb, structure.Id)
	bsv.validateDescription(tb, structure.Description)
	bsv.validateEncoding(tb, structure.Encoding)
	bsv.validateSize(tb, structure.Size)

	if bsv.validateRFCEnvelope != nil {
		require.NotNil(tb, structure.Envelope)
		bsv.validateRFCEnvelope.check(tb, structure.Envelope)
	} else {
		require.Nil(tb, structure.Envelope)
	}

	require.Equal(tb, len(structure.Parts), len(bsv.validateParts))

	for i := 0; i < len(bsv.validateParts); i++ {
		bsv.validateParts[i].check(tb, structure.Parts[i])
	}

	if bsv.validateRFCBodyStructure != nil {
		require.NotNil(tb, structure.BodyStructure)
		bsv.validateRFCBodyStructure.check(tb, structure.BodyStructure)
	} else {
		require.Nil(tb, structure.BodyStructure)
	}

	bsv.validateLines(tb, structure.Lines)
	bsv.validateExtended(tb, structure.Extended)
	bsv.validateDisposition(tb, structure.Disposition)
	bsv.validateDispositionParams(tb, structure.DispositionParams)
	bsv.validateLanguage(tb, structure.Language)
	bsv.validateLocation(tb, structure.Location)
	bsv.validateMD5(tb, structure.MD5)
}

func newEmptyBodyStructureValidator() *bodyStructureValidator {
	emptyString := func(tb testing.TB, s string) {
		require.Empty(tb, s)
	}

	return &bodyStructureValidator{
		validateMIMEType:    emptyString,
		validateMimeSubType: emptyString,
		validateParams: func(tb testing.TB, m map[string]string) {
			require.Empty(tb, m)
		},
		validateID:          emptyString,
		validateDescription: emptyString,
		validateEncoding:    emptyString,
		validateSize: func(tb testing.TB, u uint32) {
			require.Zero(tb, u)
		},
		validateParts:            nil,
		validateRFCEnvelope:      nil,
		validateRFCBodyStructure: nil,
		validateLines: func(tb testing.TB, u uint32) {
			require.Zero(tb, 0)
		},
		validateExtended: func(tb testing.TB, b bool) {
			require.True(tb, b)
		},
		validateDisposition: emptyString,
		validateDispositionParams: func(tb testing.TB, m map[string]string) {
			require.Empty(tb, m)
		},
		validateLanguage: func(tb testing.TB, i []string) {
			require.Empty(tb, i)
		},
		validateLocation: func(tb testing.TB, i []string) {
			require.Empty(tb, i)
		},
		validateMD5: emptyString,
	}
}

type bodyStructureValidatorBuilder bodyStructureValidator

func (b *bodyStructureValidatorBuilder) wantMIMEType(mimeType string) {
	b.validateMIMEType = func(tb testing.TB, s string) {
		require.Equal(tb, mimeType, s)
	}
}

func (b *bodyStructureValidatorBuilder) wantMIMESubType(mimeSubType string) {
	b.validateMimeSubType = func(tb testing.TB, s string) {
		require.Equal(tb, mimeSubType, s)
	}
}

func (b *bodyStructureValidatorBuilder) wantParams(params map[string]string) {
	b.validateParams = func(tb testing.TB, m map[string]string) {
		require.Equal(tb, params, m)
	}
}

func (b *bodyStructureValidatorBuilder) wantId(id string) {
	b.validateID = func(tb testing.TB, s string) {
		require.Equal(tb, id, s)
	}
}

func (b *bodyStructureValidatorBuilder) wantDescription(description string) {
	b.validateDescription = func(tb testing.TB, s string) {
		require.Equal(tb, description, s)
	}
}

func (b *bodyStructureValidatorBuilder) wantEncoding(encoding string) {
	b.validateEncoding = func(tb testing.TB, s string) {
		require.Equal(tb, encoding, s)
	}
}

func (b *bodyStructureValidatorBuilder) wantSize(size uint32) {
	b.validateSize = func(tb testing.TB, u uint32) {
		require.Equal(tb, size, u)
	}
}

func (b *bodyStructureValidatorBuilder) wantPart(fn func(builder *bodyStructureValidatorBuilder)) {
	builder := newEmptyBodyStructureValidator()
	fn((*bodyStructureValidatorBuilder)(builder))
	b.validateParts = append(b.validateParts, builder)
}

func (b *bodyStructureValidatorBuilder) wantRFCEnvelope(fn func(builder *envelopeValidatorBuilder)) {
	validator := newEmptyEnvelopValidator()
	builder := &envelopeValidatorBuilder{
		validator: validator,
	}
	fn(builder)

	b.validateRFCEnvelope = validator
}

func (b *bodyStructureValidatorBuilder) wantRFCBodyStructure(fn func(builder *bodyStructureValidatorBuilder)) {
	builder := newEmptyBodyStructureValidator()
	fn((*bodyStructureValidatorBuilder)(builder))
	b.validateRFCBodyStructure = builder
}

func (b *bodyStructureValidatorBuilder) wantLines(size uint32) {
	b.validateLines = func(tb testing.TB, u uint32) {
		require.Equal(tb, size, u)
	}
}

func (b *bodyStructureValidatorBuilder) wantExtended(value bool) {
	b.validateExtended = func(tb testing.TB, b bool) {
		require.Equal(tb, value, b)
	}
}

func (b *bodyStructureValidatorBuilder) wantDisposition(disposition string) {
	b.validateDisposition = func(tb testing.TB, s string) {
		require.Equal(tb, disposition, s)
	}
}

func (b *bodyStructureValidatorBuilder) wantDispositionParams(dispositionParams map[string]string) {
	b.validateDispositionParams = func(tb testing.TB, m map[string]string) {
		require.Equal(tb, dispositionParams, m)
	}
}

func (b *bodyStructureValidatorBuilder) wantLanguage(language ...string) {
	b.validateLanguage = func(tb testing.TB, i []string) {
		require.ElementsMatch(tb, language, i)
	}
}

func (b *bodyStructureValidatorBuilder) wantLocation(location ...string) {
	b.validateLocation = func(tb testing.TB, i []string) {
		require.ElementsMatch(tb, location, i)
	}
}

func (b *bodyStructureValidatorBuilder) wantMD5(md5 string) {
	b.validateMD5 = func(tb testing.TB, s string) {
		require.Equal(tb, md5, s)
	}
}
