package finality

import (
	"container/list"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/recorder"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
)

type finalityImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.RWMutex

	//============================ basic information =============================================

	// author indicates the identifier of current node.
	author uint64

	//============================ executor service processor =============================================

	//============================ executor service processor =============================================

	// streamMutex is used to process the concurrency of streams processing.
	streamMutex sync.Mutex

	// streams is used to record the committed query stream.
	streams *list.List

	// count indicates the number of query stream in streams list.
	count int64

	// closeC is used to quit the service.
	closeC chan bool

	//============================ order rule for block generation ========================================

	// cRecorder is used to record the command info.
	cRecorder api.CommandRecorder

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	// orderSeq tracks the real committed partial order sequence number.
	orderSeq map[uint64]uint64

	//============================= internal interfaces =========================================

	// reader is used to read partial orders from meta pool tracker.
	reader api.MetaReader

	// metrics is used to record the metric of current node's executor.
	metrics *metrics.ExecutorMetrics

	//============================== external interfaces ==========================================

	// logger is used to print logs.
	logger external.Logger
}

func NewFinality(conf Config) *finalityImpl {
	author := conf.Author
	orderSeq := make(map[uint64]uint64)

	for i := 0; i < conf.N; i++ {
		id := uint64(i + 1)
		orderSeq[id] = uint64(0)
	}

	cRecorder := recorder.NewCommandRecorder(author, conf.N, conf.Logger)
	return &finalityImpl{
		author:    author,
		rules:     newOrderRule(conf, cRecorder),
		cRecorder: cRecorder,
		reader:    conf.Mgr,
		logger:    conf.Logger,
		orderSeq:  orderSeq,
		streams:   list.New(),
		closeC:    make(chan bool),
		metrics:   conf.Metrics.ExecutorMetrics,
	}
}

// CommitStream is used to commit the partial order stream.
func (ei *finalityImpl) CommitStream(qStream types.QueryStream) {
	ei.streamMutex.Lock()
	ei.streams.PushBack(qStream)
	ei.streamMutex.Unlock()
	atomic.AddInt64(&ei.count, 1)
}

func (ei *finalityImpl) Run() {
	for {
		select {
		case <-ei.closeC:
			return
		default:
			ei.processStreamList()
		}
	}
}

func (ei *finalityImpl) Quit() {
	select {
	case <-ei.closeC:
	default:
		close(ei.closeC)
	}
}

func (ei *finalityImpl) processStreamList() {
	if atomic.LoadInt64(&ei.count) == 0 {
		return
	}

	ei.streamMutex.Lock()
	e := ei.streams.Front()
	atomic.AddInt64(&ei.count, -1)
	qStream := e.Value.(types.QueryStream)
	ei.streams.Remove(e)
	ei.streamMutex.Unlock()

	ei.commitStream(qStream)
}

func (ei *finalityImpl) commitStream(qStream types.QueryStream) {
	if len(qStream) == 0 {
		// nil partial order batch means we should skip the current commitment attempt.
		return
	}

	start := time.Now()

	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	partials := ei.reader.ReadPartials(qStream)

	var oStream types.OrderStream

	for _, pOrder := range partials {
		// commit metrics.
		ei.metrics.CommitPartialOrder(pOrder)

		startNo := ei.orderSeq[pOrder.Author()]

		infos, endNo := types.NewOrderInfos(startNo, pOrder)

		ei.orderSeq[pOrder.Author()] = endNo
		oStream = append(oStream, infos...)
	}
	sort.Sort(oStream)
	ei.logger.Debugf("[%d] commit order info stream len %d: %v", ei.author, len(oStream), oStream)

	updated := false
	for _, oInfo := range oStream {
		// order rule 1: collection rule, collect the partial order info.
		updated = ei.rules.collect.collectPartials(oInfo)
	}

	if updated {
		ei.rules.processPartialOrder()
	}

	// record metrics.
	ei.metrics.CommitStream(start)
}
