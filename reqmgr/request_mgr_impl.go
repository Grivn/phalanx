package reqmgr

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type requestMgrImpl struct {
	n int

	author uint64

	pool map[uint64]*requestPool

	logger external.Logger
}

func newRequestMgrImpl(n int, author uint64, replyC chan *commonProto.BatchId, logger external.Logger) *requestMgrImpl {
	logger.Noticef("[INIT] replica %d init request manager, cluster amount %d", author, n)
	rps := make(map[uint64]*requestPool)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		rp := newRequestPool(id, replyC, logger)
		rps[id] = rp
	}

	return &requestMgrImpl{
		n:      n,
		author: author,
		pool:   rps,
		logger: logger,
	}
}

func (rm *requestMgrImpl) start() {
	for _, pool := range rm.pool {
		pool.start()
	}
}

func (rm *requestMgrImpl) stop() {
	for _, pool := range rm.pool {
		pool.stop()
	}
}

func (rm *requestMgrImpl) record(msg *commonProto.OrderedMsg) {
	rm.pool[msg.Author].record(msg)
}
