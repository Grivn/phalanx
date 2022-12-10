package finality_v1

import (
	"sort"
	"sync"
	"time"

	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/service/finality_v1/finengine_v1"
	"github.com/Grivn/phalanx/pkg/utils/streamcache"
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

	// phalanxAnchor is used to generate blocks with phalanx anchor-based ordering rule.
	phalanxAnchor api.FinalityEngine

	// timestampAnchor is used to generate blocks with timestamp anchor-based ordering rule.
	timestampAnchor api.FinalityEngine

	// timestampBased is used to generate blocks with timestamp-based ordering rule.
	timestampBased api.FinalityEngine

	//============================= internal interfaces =========================================

	// reader is used to read partial orders from meta pool tracker.
	reader api.MetaReader

	// metrics is used to record the metric of current node's executor.
	metrics *metrics.ExecutorMetrics

	//============================== external interfaces ==========================================

	// logger is used to print logs.
	logger external.Logger
}

func NewFinality(
	conf config.PhalanxConf,
	meta api.MetaPool,
	executor external.Executor,
	logger external.Logger,
	ms *metrics.Metrics) *finalityImpl {
	author := conf.NodeID
	orderSeq := make(map[uint64]uint64)

	for i := 0; i < conf.NodeCount; i++ {
		id := uint64(i + 1)
		orderSeq[id] = uint64(0)
	}

	return &finalityImpl{
		author:          author,
		cache:           streamcache.NewStreamCache(),
		closeC:          make(chan bool),
		orderSeq:        orderSeq,
		phalanxAnchor:   finengine_v1.NewPhalanxAnchorBasedOrdering(conf, meta, executor, logger, ms),
		timestampAnchor: finengine_v1.NewTimestampAnchorBasedOrdering(conf, meta, executor, logger, ms),
		timestampBased:  finengine_v1.NewTimestampBasedOrdering(conf, meta, logger, ms),
		reader:          meta,
		metrics:         ms.ExecutorMetrics,
		logger:          logger,
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

	start := time.Now()
	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	var oStream types.OrderStream
	partials := ei.reader.ReadPartials(qStream)
	for _, pOrder := range partials {
		// commit metrics.
		ei.metrics.CommitPartialOrder(pOrder)

		// select the command infos we need to commit, and update the latest committed sequence number.
		startNo := ei.orderSeq[pOrder.Author()]
		infos, endNo := types.PartialOrderToOrderInfos(startNo, pOrder)
		ei.orderSeq[pOrder.Author()] = endNo

		// record the committed command infos.
		oStream = append(oStream, infos...)
	}
	sort.Sort(oStream) // sort the command infos according to generator id and sequence number.
	ei.logger.Debugf("[%d] commit order info stream len %d: %v", ei.author, len(oStream), oStream)

	ei.phalanxAnchor.CommitOrderStream(oStream)
	ei.timestampBased.CommitOrderStream(oStream)
	ei.timestampAnchor.CommitOrderStream(oStream)

	// record metrics.
	ei.metrics.CommitStream(start)
}
