package consensus

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type usigProxy struct {
	author uint64
	auth   api.Authenticator
	logger external.Logger
}

func (up *usigProxy) generateSignedMsg(msg *commonProto.OrderedMsg) *commonProto.SignedMsg {
	var (
		payload []byte
		sig     []byte
		err     error
	)

	switch msg.Type {
	case commonProto.OrderType_REQ:
		up.logger.Infof("Replica %d is trying to generate an ordered req", up.author)
		payload, err = proto.Marshal(msg)
		if err != nil {
			up.logger.Errorf("Marshal error: %s", err)
			return nil
		}
		sig, err = up.auth.GenerateMessageAuthenTag(api.ReplicaAuthen, payload)
		if err != nil {
			up.logger.Errorf("Generate authentication failed: %d", err)
			return nil
		}
	case commonProto.OrderType_LOG:
		up.logger.Infof("Replica %d is trying to generate an ordered log", up.author)
		payload, err = proto.Marshal(msg)
		if err != nil {
			up.logger.Errorf("Marshal error: %s", err)
			return nil
		}
		sig, err = up.auth.GenerateMessageAuthenTag(api.USIGAuthen, payload)
		if err != nil {
			up.logger.Errorf("Generate authentication failed: %d", err)
			return nil
		}
	default:
		up.logger.Warningf("Invalid ordered message type: code %d", msg.Type)
		return nil
	}

	return &commonProto.SignedMsg{
		Type:      commonProto.OrderType_REQ,
		Author:    up.author,
		Payload:   payload,
		Signature: sig,
	}
}

func (up *usigProxy) verifySignedMsg(signed *commonProto.SignedMsg) bool {
	var err error

	switch signed.Type {
	case commonProto.OrderType_REQ:
		up.logger.Infof("Replica %d received an ordered req, try to verify it", up.author)
		err = up.auth.VerifyMessageAuthenTag(api.ReplicaAuthen, uint32(signed.Author), signed.Payload, signed.Signature)
	case commonProto.OrderType_LOG:
		up.logger.Infof("Replica %d received an ordered log, try to verify it", up.author)
		err = up.auth.VerifyMessageAuthenTag(api.USIGAuthen, uint32(signed.Author), signed.Payload, signed.Signature)
	default:
		up.logger.Warningf("Invalid ordered message type: code %d", signed.Type)
		return false
	}

	if err != nil {
		up.logger.Errorf("Invalid ordered message: %d", err)
		return false
	}
	return true
}
