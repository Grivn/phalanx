// Code generated by MockGen. DO NOT EDIT.
// Source: executor.go

package mocks

import (
	protos "github.com/Grivn/phalanx/common/protos"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockExecutorService is a mock of ExecutorService interface
type MockExecutorService struct {
	ctrl     *gomock.Controller
	recorder *MockExecutorServiceMockRecorder
}

// MockExecutorServiceMockRecorder is the mock recorder for MockExecutorService
type MockExecutorServiceMockRecorder struct {
	mock *MockExecutorService
}

// NewMockExecutorService creates a new mock instance
func NewMockExecutorService(ctrl *gomock.Controller) *MockExecutorService {
	mock := &MockExecutorService{ctrl: ctrl}
	mock.recorder = &MockExecutorServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockExecutorService) EXPECT() *MockExecutorServiceMockRecorder {
	return _m.recorder
}

// Execute mocks base method
func (_m *MockExecutorService) Execute(txs []*protos.Transaction, seqNo uint64, timestamp int64) {
	_m.ctrl.Call(_m, "Execute", txs, seqNo, timestamp)
}

// Execute indicates an expected call of Execute
func (_mr *MockExecutorServiceMockRecorder) Execute(arg0, arg1, arg2 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCallWithMethodType(_mr.mock, "Execute", reflect.TypeOf((*MockExecutorService)(nil).Execute), arg0, arg1, arg2)
}
