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

// Authorize mocks base method.
func (m *MockConnector) Authorize(arg0, arg1 string) bool {
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

// CreateLabel mocks base method.
func (m *MockConnector) CreateLabel(arg0 context.Context, arg1 []string) (imap.Mailbox, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateLabel", arg0, arg1)
	ret0, _ := ret[0].(imap.Mailbox)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateLabel indicates an expected call of CreateLabel.
func (mr *MockConnectorMockRecorder) CreateLabel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateLabel", reflect.TypeOf((*MockConnector)(nil).CreateLabel), arg0, arg1)
}

// CreateMessage mocks base method.
func (m *MockConnector) CreateMessage(arg0 context.Context, arg1 imap.LabelID, arg2 []byte, arg3 imap.FlagSet, arg4 time.Time) (imap.Message, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateMessage", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(imap.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateMessage indicates an expected call of CreateMessage.
func (mr *MockConnectorMockRecorder) CreateMessage(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateMessage", reflect.TypeOf((*MockConnector)(nil).CreateMessage), arg0, arg1, arg2, arg3, arg4)
}

// DeleteLabel mocks base method.
func (m *MockConnector) DeleteLabel(arg0 context.Context, arg1 imap.LabelID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteLabel", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteLabel indicates an expected call of DeleteLabel.
func (mr *MockConnectorMockRecorder) DeleteLabel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteLabel", reflect.TypeOf((*MockConnector)(nil).DeleteLabel), arg0, arg1)
}

// GetLabel mocks base method.
func (m *MockConnector) GetLabel(arg0 context.Context, arg1 imap.LabelID) (imap.Mailbox, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLabel", arg0, arg1)
	ret0, _ := ret[0].(imap.Mailbox)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLabel indicates an expected call of GetLabel.
func (mr *MockConnectorMockRecorder) GetLabel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLabel", reflect.TypeOf((*MockConnector)(nil).GetLabel), arg0, arg1)
}

// GetMessage mocks base method.
func (m *MockConnector) GetMessage(arg0 context.Context, arg1 imap.MessageID) (imap.Message, []imap.LabelID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMessage", arg0, arg1)
	ret0, _ := ret[0].(imap.Message)
	ret1, _ := ret[1].([]imap.LabelID)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetMessage indicates an expected call of GetMessage.
func (mr *MockConnectorMockRecorder) GetMessage(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMessage", reflect.TypeOf((*MockConnector)(nil).GetMessage), arg0, arg1)
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

// LabelMessages mocks base method.
func (m *MockConnector) LabelMessages(arg0 context.Context, arg1 []imap.MessageID, arg2 imap.LabelID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LabelMessages", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// LabelMessages indicates an expected call of LabelMessages.
func (mr *MockConnectorMockRecorder) LabelMessages(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LabelMessages", reflect.TypeOf((*MockConnector)(nil).LabelMessages), arg0, arg1, arg2)
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
func (m *MockConnector) MoveMessages(arg0 context.Context, arg1 []imap.MessageID, arg2, arg3 imap.LabelID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MoveMessages", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// MoveMessages indicates an expected call of MoveMessages.
func (mr *MockConnectorMockRecorder) MoveMessages(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MoveMessages", reflect.TypeOf((*MockConnector)(nil).MoveMessages), arg0, arg1, arg2, arg3)
}

// UnlabelMessages mocks base method.
func (m *MockConnector) UnlabelMessages(arg0 context.Context, arg1 []imap.MessageID, arg2 imap.LabelID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnlabelMessages", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UnlabelMessages indicates an expected call of UnlabelMessages.
func (mr *MockConnectorMockRecorder) UnlabelMessages(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnlabelMessages", reflect.TypeOf((*MockConnector)(nil).UnlabelMessages), arg0, arg1, arg2)
}

// UpdateLabel mocks base method.
func (m *MockConnector) UpdateLabel(arg0 context.Context, arg1 imap.LabelID, arg2 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateLabel", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateLabel indicates an expected call of UpdateLabel.
func (mr *MockConnectorMockRecorder) UpdateLabel(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLabel", reflect.TypeOf((*MockConnector)(nil).UpdateLabel), arg0, arg1, arg2)
}

// ValidateCreate mocks base method.
func (m *MockConnector) ValidateCreate(arg0 []string) (imap.FlagSet, imap.FlagSet, imap.FlagSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateCreate", arg0)
	ret0, _ := ret[0].(imap.FlagSet)
	ret1, _ := ret[1].(imap.FlagSet)
	ret2, _ := ret[2].(imap.FlagSet)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ValidateCreate indicates an expected call of ValidateCreate.
func (mr *MockConnectorMockRecorder) ValidateCreate(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateCreate", reflect.TypeOf((*MockConnector)(nil).ValidateCreate), arg0)
}

// ValidateDelete mocks base method.
func (m *MockConnector) ValidateDelete(arg0 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateDelete", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateDelete indicates an expected call of ValidateDelete.
func (mr *MockConnectorMockRecorder) ValidateDelete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateDelete", reflect.TypeOf((*MockConnector)(nil).ValidateDelete), arg0)
}

// ValidateUpdate mocks base method.
func (m *MockConnector) ValidateUpdate(arg0, arg1 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateUpdate", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateUpdate indicates an expected call of ValidateUpdate.
func (mr *MockConnectorMockRecorder) ValidateUpdate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateUpdate", reflect.TypeOf((*MockConnector)(nil).ValidateUpdate), arg0, arg1)
}
