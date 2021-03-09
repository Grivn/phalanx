package consensus

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type collectorMgr struct {
	author uint64

	recvC chan *commonProto.SignedMsg

	close chan bool

	logger external.Logger
}

func newCollectMgr() *collectorMgr {
	return &collectorMgr{}
}

func (cm *collectorMgr) start() {
	go cm.listener()
}

func (cm *collectorMgr) stop() {
	close(cm.close)
}

func (cm *collectorMgr) listener() {
	for {
		select {
		case <-cm.close:
			cm.logger.Notice("exist request listener")
			return
		case req := <-cm.recvC:
			cm.processOrderedReq(req)
		}
	}
}

func (cm *collectorMgr) processOrderedReq(signed *commonProto.SignedMsg) {
	switch signed.Type {
	case commonProto.OrderType_REQ:
	case commonProto.OrderType_LOG:
	default:
		return
	}
}
