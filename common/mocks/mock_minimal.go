package mocks

import "github.com/golang/mock/gomock"

func NewMockMinimalNetworkService(ctrl *gomock.Controller) *MockNetworkService {
	mock := NewMockNetworkService(ctrl)
	mock.EXPECT().Broadcast(gomock.Any()).AnyTimes()
	mock.EXPECT().Unicast(gomock.Any()).AnyTimes()
	return mock
}

func NewMockMinimalExecutorService(ctrl *gomock.Controller) *MockExecutorService {
	mock := NewMockExecutorService(ctrl)
	mock.EXPECT().Execute(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()
	return mock
}

func NewMockMinimalLogger(ctrl *gomock.Controller) *MockLogger {
	mock := NewMockLogger(ctrl)
	mock.EXPECT().Debug(gomock.Any()).AnyTimes()
	mock.EXPECT().Debugf(gomock.Any()).AnyTimes()
	mock.EXPECT().Info(gomock.Any()).AnyTimes()
	mock.EXPECT().Infof(gomock.Any()).AnyTimes()
	mock.EXPECT().Notice(gomock.Any()).AnyTimes()
	mock.EXPECT().Noticef(gomock.Any()).AnyTimes()
	mock.EXPECT().Warning(gomock.Any()).AnyTimes()
	mock.EXPECT().Warningf(gomock.Any()).AnyTimes()
	mock.EXPECT().Error(gomock.Any()).AnyTimes()
	mock.EXPECT().Errorf(gomock.Any()).AnyTimes()
	mock.EXPECT().Critical(gomock.Any()).AnyTimes()
	mock.EXPECT().Criticalf(gomock.Any()).AnyTimes()
	return mock
}
