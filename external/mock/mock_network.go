// Code generated by MockGen. DO NOT EDIT.
// Source: ../network.go

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
func (_m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return _m.recorder
}

// Broadcast mocks base method
func (_m *MockNetwork) Broadcast(message *protos.ConsensusMessage) {
	_m.ctrl.Call(_m, "Broadcast", message)
}

// Broadcast indicates an expected call of Broadcast
func (_mr *MockNetworkMockRecorder) Broadcast(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Broadcast", reflect.TypeOf((*MockNetwork)(nil).Broadcast), arg0)
}

// Unicast mocks base method
func (_m *MockNetwork) Unicast(message *protos.ConsensusMessage) {
	_m.ctrl.Call(_m, "Unicast", message)
}

// Unicast indicates an expected call of Unicast
func (_mr *MockNetworkMockRecorder) Unicast(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Unicast", reflect.TypeOf((*MockNetwork)(nil).Unicast), arg0)
}
