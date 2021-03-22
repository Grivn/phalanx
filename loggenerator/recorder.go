package loggenerator

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type recorder struct {
	author uint64
	id     uint64
	agreed map[uint64]bool
	logs   map[uint64]*commonProto.OrderedMsg
	logger external.Logger
}

func newRecorder(author, id uint64, logger external.Logger) *recorder {
	logger.Noticef("replica %d init recorder for replica %d", author, id)
	return &recorder{
		author: author,
		id:     id,
		agreed: make(map[uint64]bool),
		logs:   make(map[uint64]*commonProto.OrderedMsg),
		logger: logger,
	}
}

func (re *recorder) update(msg *commonProto.OrderedMsg) bool{
	if re.agreed[msg.Sequence] {
		re.logger.Warningf("[TIMEOUT] receive a timeout ordered log from replica %d with sequence %d", msg.Author, msg.Sequence)
		return false
	}

	if _, ok := re.logs[msg.Sequence]; ok {
		re.logger.Warningf("[BRANCH] replica %d received a duplicated ordered log from replica %d for sequence %d", re.author, msg.Author, msg.Sequence)
		return false
	}

	re.logs[msg.Sequence] = msg
	re.logger.Infof("[Record] replica %d receive a signed ordered log from replica %d for sequence %d", re.author, msg.Author, msg.Sequence)
	return true
}

func (re *recorder) upgrade(agreed uint64) {
	re.logger.Debugf("[UPGRADE] replica %d already agree the sequence %d, remove the logs of replica %d", re.author, agreed, re.id)
	delete(re.logs, agreed)
	re.agreed[agreed] = true
}

