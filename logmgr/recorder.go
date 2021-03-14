package logmgr

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"

	"github.com/gogo/protobuf/proto"
)

type recorder struct {
	author uint64
	id     uint64
	seqNo  uint64
	logs   map[uint64]*commonProto.OrderedMsg
	auth   api.Authenticator
	logger external.Logger
}

func newRecorder(author, id uint64, auth api.Authenticator, logger external.Logger) *recorder {
	return &recorder{
		author: author,
		id:     id,
		seqNo:  uint64(1),
		logs:   make(map[uint64]*commonProto.OrderedMsg),
		auth:   auth,
		logger: logger,
	}
}

func (re *recorder) update(signed *commonProto.SignedMsg) *commonProto.OrderedMsg {
	re.logger.Infof("[Record] replica %d receive a signed ordered log from replica %d", re.author, signed.Author)

	err := re.auth.VerifyMessageAuthenTag(api.USIGAuthen, uint32(signed.Author-1), signed.Payload, signed.Signature)
	if err != nil {
		re.logger.Warningf("[Invalid] receive an invalid ordered log from replica %d, error %s", signed.Author, err)
		return nil
	}

	msg := &commonProto.OrderedMsg{}
	_ = proto.Unmarshal(signed.Payload, msg)

	if msg.Sequence < re.seqNo {
		re.logger.Infof("[TIMEOUT] receive a timeout ordered log from replica %d with sequence %d", msg.Author, msg.Sequence)
		return nil
	}

	re.logs[msg.Sequence] = msg
	return msg
}

func (re *recorder) upgrade(agreed uint64) {
	for {
		if re.seqNo > agreed {
			break
		}

		re.logger.Debugf("[UPGRADE] replica %d already agree the sequence %d, remove the logs of replica %d", re.author, re.seqNo, re.id)
		delete(re.logs, re.seqNo)
		re.seqNo++
	}
}

