package mocks

import "github.com/Grivn/phalanx/api"

type AuthMock struct {

}

func NewAuthMock() api.Authenticator {
	return &AuthMock{}
}

func (au *AuthMock) VerifyMessageAuthenTag(role api.AuthenticationRole, id uint32, msg []byte, tag []byte) error {
	return nil
}

func (au *AuthMock) GenerateMessageAuthenTag(role api.AuthenticationRole, msg []byte) ([]byte, error) {
	return []byte("test"), nil
}
