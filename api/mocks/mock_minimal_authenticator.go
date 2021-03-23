package mocks

import "github.com/golang/mock/gomock"

func NewAuthenticatorMinimal(ctrl *gomock.Controller) *MockAuthenticator {
	ma := NewMockAuthenticator(ctrl)
	ma.EXPECT().GenerateMessageAuthenTag(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	ma.EXPECT().VerifyMessageAuthenTag(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	return ma
}
