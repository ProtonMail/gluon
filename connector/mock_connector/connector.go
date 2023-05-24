// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ProtonMail/gluon/connector (interfaces: Connector)

// Package mock_connector is a generated GoMock package.
package mock_connector

import (
	context "context"
	reflect "reflect"
	time "time"

	imap "github.com/ProtonMail/gluon/imap"
	gomock "github.com/golang/mock/gomock"
)

// MockConnector is a mock of Connector interface.
type MockConnector struct {
	ctrl     *gomock.Controller
	recorder *MockConnectorMockRecorder
}

// MockConnectorMockRecorder is the mock recorder for MockConnector.
type MockConnectorMockRecorder struct {
	mock *MockConnector
}

// NewMockConnector creates a new mock instance.
func NewMockConnector(ctrl *gomock.Controller) *MockConnector {
	mock := &MockConnector{ctrl: ctrl}
	mock.recorder = &MockConnectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnector) EXPECT() *MockConnectorMockRecorder {
	return m.recorder
}

// AddMessagesToMailbox mocks base method.
func (m *MockConnector) AddMessagesToMailbox(arg0 context.Context, arg1 []imap.MessageID, arg2 imap.MailboxID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddMessagesToMailbox", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddMessagesToMailbox indicates an expected call of AddMessagesToMailbox.
func (mr *MockConnectorMockRecorder) AddMessagesToMailbox(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMessagesToMailbox", reflect.TypeOf((*MockConnector)(nil).AddMessagesToMailbox), arg0, arg1, arg2)
}

// Authorize mocks base method.
func (m *MockConnector) Authorize(arg0 string, arg1 []byte) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Authorize", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// Authorize indicates an expected call of Authorize.
func (mr *MockConnectorMockRecorder) Authorize(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Authorize", reflect.TypeOf((*MockConnector)(nil).Authorize), arg0, arg1)
}

// Close mocks base method.
func (m *MockConnector) Close(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockConnectorMockRecorder) Close(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockConnector)(nil).Close), arg0)
}

// CreateMailbox mocks base method.
func (m *MockConnector) CreateMailbox(arg0 context.Context, arg1 []string) (imap.Mailbox, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMailbox", arg0, arg1)
	ret0, _ := ret[0].(imap.Mailbox)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMailbox indicates an expected call of CreateMailbox.
func (mr *MockConnectorMockRecorder) CreateMailbox(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMailbox", reflect.TypeOf((*MockConnector)(nil).CreateMailbox), arg0, arg1)
}

// CreateMessage mocks base method.
func (m *MockConnector) CreateMessage(arg0 context.Context, arg1 imap.MailboxID, arg2 []byte, arg3 imap.FlagSet, arg4 time.Time) (imap.Message, []byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMessage", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(imap.Message)
	ret1, _ := ret[1].([]byte)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateMessage indicates an expected call of CreateMessage.
func (mr *MockConnectorMockRecorder) CreateMessage(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMessage", reflect.TypeOf((*MockConnector)(nil).CreateMessage), arg0, arg1, arg2, arg3, arg4)
}

// DeleteMailbox mocks base method.
func (m *MockConnector) DeleteMailbox(arg0 context.Context, arg1 imap.MailboxID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteMailbox", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteMailbox indicates an expected call of DeleteMailbox.
func (mr *MockConnectorMockRecorder) DeleteMailbox(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMailbox", reflect.TypeOf((*MockConnector)(nil).DeleteMailbox), arg0, arg1)
}

// GetMailboxVisibility mocks base method.
func (m *MockConnector) GetMailboxVisibility(arg0 context.Context, arg1 imap.MailboxID) imap.MailboxVisibility {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMailboxVisibility", arg0, arg1)
	ret0, _ := ret[0].(imap.MailboxVisibility)
	return ret0
}

// GetMailboxVisibility indicates an expected call of GetMailboxVisibility.
func (mr *MockConnectorMockRecorder) GetMailboxVisibility(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMailboxVisibility", reflect.TypeOf((*MockConnector)(nil).GetMailboxVisibility), arg0, arg1)
}

// GetMessageLiteral mocks base method.
func (m *MockConnector) GetMessageLiteral(arg0 context.Context, arg1 imap.MessageID) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMessageLiteral", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMessageLiteral indicates an expected call of GetMessageLiteral.
func (mr *MockConnectorMockRecorder) GetMessageLiteral(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMessageLiteral", reflect.TypeOf((*MockConnector)(nil).GetMessageLiteral), arg0, arg1)
}

// GetUpdates mocks base method.
func (m *MockConnector) GetUpdates() <-chan imap.Update {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUpdates")
	ret0, _ := ret[0].(<-chan imap.Update)
	return ret0
}

// GetUpdates indicates an expected call of GetUpdates.
func (mr *MockConnectorMockRecorder) GetUpdates() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUpdates", reflect.TypeOf((*MockConnector)(nil).GetUpdates))
}

// MarkMessagesFlagged mocks base method.
func (m *MockConnector) MarkMessagesFlagged(arg0 context.Context, arg1 []imap.MessageID, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkMessagesFlagged", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkMessagesFlagged indicates an expected call of MarkMessagesFlagged.
func (mr *MockConnectorMockRecorder) MarkMessagesFlagged(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkMessagesFlagged", reflect.TypeOf((*MockConnector)(nil).MarkMessagesFlagged), arg0, arg1, arg2)
}

// MarkMessagesSeen mocks base method.
func (m *MockConnector) MarkMessagesSeen(arg0 context.Context, arg1 []imap.MessageID, arg2 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkMessagesSeen", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// MarkMessagesSeen indicates an expected call of MarkMessagesSeen.
func (mr *MockConnectorMockRecorder) MarkMessagesSeen(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkMessagesSeen", reflect.TypeOf((*MockConnector)(nil).MarkMessagesSeen), arg0, arg1, arg2)
}

// MoveMessages mocks base method.
func (m *MockConnector) MoveMessages(arg0 context.Context, arg1 []imap.MessageID, arg2, arg3 imap.MailboxID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MoveMessages", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// MoveMessages indicates an expected call of MoveMessages.
func (mr *MockConnectorMockRecorder) MoveMessages(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveMessages", reflect.TypeOf((*MockConnector)(nil).MoveMessages), arg0, arg1, arg2, arg3)
}

// RemoveMessagesFromMailbox mocks base method.
func (m *MockConnector) RemoveMessagesFromMailbox(arg0 context.Context, arg1 []imap.MessageID, arg2 imap.MailboxID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveMessagesFromMailbox", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveMessagesFromMailbox indicates an expected call of RemoveMessagesFromMailbox.
func (mr *MockConnectorMockRecorder) RemoveMessagesFromMailbox(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveMessagesFromMailbox", reflect.TypeOf((*MockConnector)(nil).RemoveMessagesFromMailbox), arg0, arg1, arg2)
}

// UpdateMailboxName mocks base method.
func (m *MockConnector) UpdateMailboxName(arg0 context.Context, arg1 imap.MailboxID, arg2 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateMailboxName", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateMailboxName indicates an expected call of UpdateMailboxName.
func (mr *MockConnectorMockRecorder) UpdateMailboxName(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMailboxName", reflect.TypeOf((*MockConnector)(nil).UpdateMailboxName), arg0, arg1, arg2)
}
