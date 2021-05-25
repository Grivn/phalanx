// Code generated by MockGen. DO NOT EDIT.
// Source: ../network.go

// Package mocks is a generated GoMock package.
package mocks

import (
	protos "github.com/Grivn/phalanx/common/protos"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockNetwork is a mock of Network interface
type MockNetwork struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkMockRecorder
}

// MockNetworkMockRecorder is the mock recorder for MockNetwork
type MockNetworkMockRecorder struct {
	mock *MockNetwork
}

// NewMockNetwork creates a new mock instance
func NewMockNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := &MockNetwork{ctrl: ctrl}
	mock.recorder = &MockNetworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return m.recorder
}

// Broadcast mocks base method
func (m *MockNetwork) Broadcast(msg *protos.CommMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Broadcast", msg)
}

// Broadcast indicates an expected call of Broadcast
func (mr *MockNetworkMockRecorder) Broadcast(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Broadcast", reflect.TypeOf((*MockNetwork)(nil).Broadcast), msg)
}
