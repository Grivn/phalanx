package finality

import (
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/utils/streamcache"
	"sort"
	"sync"
)

type finalityImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.RWMutex

	//============================ basic information =============================================

	// author indicates the identifier of current node.
	author uint64

	//========================= concurrency committed query stream processor =============================

	// cache is used to process the query streams commit into finality module.
	cache api.StreamCache

	// closeC is used to quit the service.
	closeC chan bool

	//============================ order rule for block generation ========================================

	// orderSeq tracks the real committed partial order sequence number.
	orderSeq map[uint64]uint64

	// engine is used to generate blocks with finality ordering strategy.
	engine api.FinalityEngine

	//============================= internal interfaces =========================================

	attemptTracker api.AttemptTracker

	//============================== external interfaces ==========================================

	// logger is used to print logs.
	logger external.Logger
}

func NewFinality(
	conf config.PhalanxConf,
	attemptTracker api.AttemptTracker,
	engine api.FinalityEngine,
	logger external.Logger) *finalityImpl {
	author := conf.NodeID
	orderSeq := make(map[uint64]uint64)

	for i := 0; i < conf.NodeCount; i++ {
		id := uint64(i + 1)
		orderSeq[id] = uint64(0)
	}

	return &finalityImpl{
		author:         author,
		cache:          streamcache.NewStreamCache(),
		closeC:         make(chan bool),
		orderSeq:       orderSeq,
		engine:         engine,
		attemptTracker: attemptTracker,
		logger:         logger,
	}
}

// CommitStream is used to commit the partial order stream.
func (ei *finalityImpl) CommitStream(qStream types.QueryStream) {
	if len(qStream) == 0 {
		return
	}
	ei.cache.Append(qStream)
}

func (ei *finalityImpl) Run() {
	go ei.listener()
}

func (ei *finalityImpl) Quit() {
	select {
	case <-ei.closeC:
	default:
		close(ei.closeC)
	}
}

func (ei *finalityImpl) listener() {
	for {
		select {
		case <-ei.closeC:
			return
		default:
			ei.processStreamList()
		}
	}
}

func (ei *finalityImpl) processStreamList() {
	item := ei.cache.Front()
	qStream, ok := item.(types.QueryStream)
	if !ok {
		return
	}
	ei.commitStream(qStream)
}

func (ei *finalityImpl) commitStream(qStream types.QueryStream) {
	if len(qStream) == 0 {
		// nil partial order batch means we should skip the current commitment attempt.
		return
	}

	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	var oStream types.OrderStream
	attempts := ei.readOrderAttempts(qStream)
	for _, attempt := range attempts {
		// select the command infos we need to commit, and update the latest committed sequence number.
		startNo := ei.orderSeq[attempt.NodeID]
		infos, endNo := types.OrderAttemptToOrderInfos(startNo, attempt)
		ei.orderSeq[attempt.NodeID] = endNo

		// record the committed command infos.
		oStream = append(oStream, infos...)
	}
	sort.Sort(oStream) // sort the command infos according to generator id and sequence number.
	ei.logger.Debugf("[%d] commit order info stream len %d: %v", ei.author, len(oStream), oStream)

	ei.engine.CommitOrderStream(oStream)
}

func (ei *finalityImpl) readOrderAttempts(qStream types.QueryStream) []*protos.OrderAttempt {
	var res []*protos.OrderAttempt

	for _, qIndex := range qStream {
		attempt := ei.attemptTracker.Get(qIndex)

		for {
			if attempt != nil {
				break
			}

			// if we could not read the partial order, just try the next time.
			attempt = ei.attemptTracker.Get(qIndex)
		}

		res = append(res, attempt)
	}

	return res
}
