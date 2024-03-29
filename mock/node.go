// Code generated by MockGen. DO NOT EDIT.
// Source: node.go

// Package mock is a generated GoMock package.
package mock

import (
	v1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	gomock "github.com/golang/mock/gomock"
	routing "github.com/qiangxue/fasthttp-routing"
	reflect "reflect"
)

// MockNode is a mock of Node interface.
type MockNode struct {
	ctrl     *gomock.Controller
	recorder *MockNodeMockRecorder
}

// MockNodeMockRecorder is the mock recorder for MockNode.
type MockNodeMockRecorder struct {
	mock *MockNode
}

// NewMockNode creates a new mock instance.
func NewMockNode(ctrl *gomock.Controller) *MockNode {
	mock := &MockNode{ctrl: ctrl}
	mock.recorder = &MockNodeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockNode) EXPECT() *MockNodeMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockNode) Get() (*v1.Node, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get")
	ret0, _ := ret[0].(*v1.Node)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockNodeMockRecorder) Get() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockNode)(nil).Get))
}

// Desire mocks base method.
func (m *MockNode) Desire(desired v1.Desire, override bool) (v1.Delta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Desire", desired, override)
	ret0, _ := ret[0].(v1.Delta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Desire indicates an expected call of Desire.
func (mr *MockNodeMockRecorder) Desire(desired, override interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Desire", reflect.TypeOf((*MockNode)(nil).Desire), desired, override)
}

// Report mocks base method.
func (m *MockNode) Report(reported v1.Report, override bool) (v1.Delta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Report", reported, override)
	ret0, _ := ret[0].(v1.Delta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Report indicates an expected call of Report.
func (mr *MockNodeMockRecorder) Report(reported, override interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Report", reflect.TypeOf((*MockNode)(nil).Report), reported, override)
}

// GetStats mocks base method.
func (m *MockNode) GetStats(ctx *routing.Context) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetStats", ctx)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetStats indicates an expected call of GetStats.
func (mr *MockNodeMockRecorder) GetStats(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetStats", reflect.TypeOf((*MockNode)(nil).GetStats), ctx)
}

// GetNodeProperties mocks base method.
func (m *MockNode) GetNodeProperties(ctx *routing.Context) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodeProperties", ctx)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNodeProperties indicates an expected call of GetNodeProperties.
func (mr *MockNodeMockRecorder) GetNodeProperties(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodeProperties", reflect.TypeOf((*MockNode)(nil).GetNodeProperties), ctx)
}

// UpdateNodeProperties mocks base method.
func (m *MockNode) UpdateNodeProperties(ctx *routing.Context) (interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateNodeProperties", ctx)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateNodeProperties indicates an expected call of UpdateNodeProperties.
func (mr *MockNodeMockRecorder) UpdateNodeProperties(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNodeProperties", reflect.TypeOf((*MockNode)(nil).UpdateNodeProperties), ctx)
}
