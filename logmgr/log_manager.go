package logmgr

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/logmgr/types"
)

func NewLogManager(n int, author uint64, replyC chan types.ReplyEvent, auth api.Authenticator, network external.Network, logger external.Logger) api.LogManager {
	return newLogMgrImpl(n, author, replyC, auth, network, logger)
}

func (lm *logMgrImpl) Start() {
	lm.start()
}

func (lm *logMgrImpl) Stop() {
	lm.stop()
}

func (lm *logMgrImpl) Generate(bid *commonProto.BatchId) {
	lm.generate(bid)
}

func (lm *logMgrImpl) Record(msg *commonProto.SignedMsg) {
	lm.record(msg)
}

func (lm *logMgrImpl) Ready(binarySet []byte) {
	lm.ready(binarySet)
}
