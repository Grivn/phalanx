package mocks

import "github.com/golang/mock/gomock"

func NewMockMinimalNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := NewMockNetwork(ctrl)
	mock.EXPECT().BroadcastPreOrder(gomock.Any()).AnyTimes()
	mock.EXPECT().SendVote(gomock.Any(), gomock.Any()).AnyTimes()
	mock.EXPECT().BroadcastQC(gomock.Any()).AnyTimes()
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
