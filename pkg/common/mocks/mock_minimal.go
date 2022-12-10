package mocks

import (
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/golang/mock/gomock"
)

func NewMockMinimalSender(ctrl *gomock.Controller) *MockSender {
	mock := NewMockSender(ctrl)
	mock.EXPECT().BroadcastPCM(gomock.Any()).AnyTimes()
	mock.EXPECT().UnicastPCM(gomock.Any()).AnyTimes()
	mock.EXPECT().BroadcastCommand(gomock.Any()).AnyTimes()
	return mock
}

func NewMockMinimalExecutor(ctrl *gomock.Controller) *MockExecutor {
	mock := NewMockExecutor(ctrl)
	mock.EXPECT().CommandExecution(gomock.Any(), gomock.Any()).AnyTimes()
	return mock
}

func NewMockMinimalLogger(ctrl *gomock.Controller) *MockLogger {
	mock := NewMockLogger(ctrl)
	mock.EXPECT().Debug(gomock.Any()).AnyTimes()
	mock.EXPECT().Debugf(gomock.Any()).AnyTimes()
	mock.EXPECT().Info(gomock.Any()).AnyTimes()
	mock.EXPECT().Infof(gomock.Any()).AnyTimes()
	mock.EXPECT().Error(gomock.Any()).AnyTimes()
	mock.EXPECT().Errorf(gomock.Any()).AnyTimes()
	return mock
}

func NewMockMinimalPrivateKey(ctrl *gomock.Controller) *MockPrivateKey {
	mock := NewMockPrivateKey(ctrl)
	mock.EXPECT().PublicKey().Return(nil).AnyTimes()
	mock.EXPECT().Algorithm().Return("mock_algo").AnyTimes()
	mock.EXPECT().Sign(gomock.Any()).Return(&protos.Certification{Signatures: nil}, nil).AnyTimes()
	return mock
}

func NewMockMinimalPublicKeys(ctrl *gomock.Controller) *MockPublicKeys {
	mock := NewMockPublicKeys(ctrl)
	mock.EXPECT().Algorithm().Return("mock_algo").AnyTimes()
	mock.EXPECT().Verify(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	return mock
}

func NewMockMinimalPublicKey(ctrl *gomock.Controller) *MockPublicKey {
	mock := NewMockPublicKey(ctrl)
	mock.EXPECT().Algorithm().Return("mock_algo").AnyTimes()
	mock.EXPECT().Verify(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	return mock
}
