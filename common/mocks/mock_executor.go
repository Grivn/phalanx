// Code generated by MockGen. DO NOT EDIT.
// Source: executor.go

package mocks

import (
	protos "github.com/Grivn/phalanx/common/protos"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockExecutionService is a mock of ExecutionService interface
type MockExecutionService struct {
	ctrl     *gomock.Controller
	recorder *MockExecutionServiceMockRecorder
}

// MockExecutionServiceMockRecorder is the mock recorder for MockExecutionService
type MockExecutionServiceMockRecorder struct {
	mock *MockExecutionService
}

// NewMockExecutionService creates a new mock instance
func NewMockExecutionService(ctrl *gomock.Controller) *MockExecutionService {
	mock := &MockExecutionService{ctrl: ctrl}
	mock.recorder = &MockExecutionServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockExecutionService) EXPECT() *MockExecutionServiceMockRecorder {
	return _m.recorder
}

// CommandExecution mocks base method
func (_m *MockExecutionService) CommandExecution(command *protos.Command, seqNo uint64, timestamp int64) {
	_m.ctrl.Call(_m, "CommandExecution", command, seqNo, timestamp)
}

// CommandExecution indicates an expected call of CommandExecution
func (_mr *MockExecutionServiceMockRecorder) CommandExecution(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "CommandExecution", reflect.TypeOf((*MockExecutionService)(nil).CommandExecution), arg0, arg1, arg2)
}
