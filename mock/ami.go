// Code generated by MockGen. DO NOT EDIT.
// Source: ami.go

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	io "io"
	reflect "reflect"

	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	ami "github.com/baetyl/baetyl/v2/ami"
	gomock "github.com/golang/mock/gomock"
)

// MockAMI is a mock of AMI interface.
type MockAMI struct {
	ctrl     *gomock.Controller
	recorder *MockAMIMockRecorder
}

// MockAMIMockRecorder is the mock recorder for MockAMI.
type MockAMIMockRecorder struct {
	mock *MockAMI
}

// NewMockAMI creates a new mock instance.
func NewMockAMI(ctrl *gomock.Controller) *MockAMI {
	mock := &MockAMI{ctrl: ctrl}
	mock.recorder = &MockAMIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAMI) EXPECT() *MockAMIMockRecorder {
	return m.recorder
}

// ApplyApp mocks base method.
func (m *MockAMI) ApplyApp(arg0 string, arg1 v1.Application, arg2 map[string]v1.Configuration, arg3 map[string]v1.Secret) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyApp", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyApp indicates an expected call of ApplyApp.
func (mr *MockAMIMockRecorder) ApplyApp(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyApp", reflect.TypeOf((*MockAMI)(nil).ApplyApp), arg0, arg1, arg2, arg3)
}

// CollectNodeInfo mocks base method.
func (m *MockAMI) CollectNodeInfo() (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CollectNodeInfo")
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CollectNodeInfo indicates an expected call of CollectNodeInfo.
func (mr *MockAMIMockRecorder) CollectNodeInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectNodeInfo", reflect.TypeOf((*MockAMI)(nil).CollectNodeInfo))
}

// CollectNodeStats mocks base method.
func (m *MockAMI) CollectNodeStats() (map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CollectNodeStats")
	ret0, _ := ret[0].(map[string]interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CollectNodeStats indicates an expected call of CollectNodeStats.
func (mr *MockAMIMockRecorder) CollectNodeStats() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CollectNodeStats", reflect.TypeOf((*MockAMI)(nil).CollectNodeStats))
}

// DeleteApp mocks base method.
func (m *MockAMI) DeleteApp(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteApp", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteApp indicates an expected call of DeleteApp.
func (mr *MockAMIMockRecorder) DeleteApp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteApp", reflect.TypeOf((*MockAMI)(nil).DeleteApp), arg0, arg1)
}

// FetchLog mocks base method.
func (m *MockAMI) FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FetchLog", namespace, service, tailLines, sinceSeconds)
	ret0, _ := ret[0].(io.ReadCloser)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FetchLog indicates an expected call of FetchLog.
func (mr *MockAMIMockRecorder) FetchLog(namespace, service, tailLines, sinceSeconds interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FetchLog", reflect.TypeOf((*MockAMI)(nil).FetchLog), namespace, service, tailLines, sinceSeconds)
}

// GetModeInfo mocks base method.
func (m *MockAMI) GetModeInfo() (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetModeInfo")
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetModeInfo indicates an expected call of GetModeInfo.
func (mr *MockAMIMockRecorder) GetModeInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetModeInfo", reflect.TypeOf((*MockAMI)(nil).GetModeInfo))
}

// RPCApp mocks base method.
func (m *MockAMI) RPCApp(url string, req *v1.RPCRequest) (*v1.RPCResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RPCApp", url, req)
	ret0, _ := ret[0].(*v1.RPCResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RPCApp indicates an expected call of RPCApp.
func (mr *MockAMIMockRecorder) RPCApp(url, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RPCApp", reflect.TypeOf((*MockAMI)(nil).RPCApp), url, req)
}

// RemoteCommand mocks base method.
func (m *MockAMI) RemoteCommand(option *ami.DebugOptions, pipe ami.Pipe) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoteCommand", option, pipe)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoteCommand indicates an expected call of RemoteCommand.
func (mr *MockAMIMockRecorder) RemoteCommand(option, pipe interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteCommand", reflect.TypeOf((*MockAMI)(nil).RemoteCommand), option, pipe)
}

// RemoteDescribe mocks base method.
func (m *MockAMI) RemoteDescribe(tp, ns, n string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoteDescribe", tp, ns, n)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RemoteDescribe indicates an expected call of RemoteDescribe.
func (mr *MockAMIMockRecorder) RemoteDescribe(tp, ns, n interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteDescribe", reflect.TypeOf((*MockAMI)(nil).RemoteDescribe), tp, ns, n)
}

// RemoteLogs mocks base method.
func (m *MockAMI) RemoteLogs(option *ami.LogsOptions, pipe ami.Pipe) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoteLogs", option, pipe)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoteLogs indicates an expected call of RemoteLogs.
func (mr *MockAMIMockRecorder) RemoteLogs(option, pipe interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteLogs", reflect.TypeOf((*MockAMI)(nil).RemoteLogs), option, pipe)
}

// RemoteWebsocket mocks base method.
func (m *MockAMI) RemoteWebsocket(ctx context.Context, option *ami.DebugOptions, pipe ami.Pipe) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoteWebsocket", ctx, option, pipe)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoteWebsocket indicates an expected call of RemoteWebsocket.
func (mr *MockAMIMockRecorder) RemoteWebsocket(ctx, option, pipe interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoteWebsocket", reflect.TypeOf((*MockAMI)(nil).RemoteWebsocket), ctx, option, pipe)
}

// StatsApps mocks base method.
func (m *MockAMI) StatsApps(arg0 string) ([]v1.AppStats, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StatsApps", arg0)
	ret0, _ := ret[0].([]v1.AppStats)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// StatsApps indicates an expected call of StatsApps.
func (mr *MockAMIMockRecorder) StatsApps(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StatsApps", reflect.TypeOf((*MockAMI)(nil).StatsApps), arg0)
}

// UpdateNodeLabels mocks base method.
func (m *MockAMI) UpdateNodeLabels(arg0 string, arg1 map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNodeLabels", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNodeLabels indicates an expected call of UpdateNodeLabels.
func (mr *MockAMIMockRecorder) UpdateNodeLabels(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNodeLabels", reflect.TypeOf((*MockAMI)(nil).UpdateNodeLabels), arg0, arg1)
}
