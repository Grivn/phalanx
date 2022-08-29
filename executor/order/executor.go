package order

import (
	"container/list"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/recorder"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
)

type executorImpl struct {
	// mutex is used to deal with the concurrent problems of executor.
	mutex sync.RWMutex

	//============================ basic information =============================================

	// author indicates the identifier of current node.
	author uint64

	//============================ order rule for block generation ========================================

	// cRecorder is used to record the command info.
	cRecorder internal.CommandRecorder

	// rules is used to generate blocks with phalanx order-rule.
	rules *orderRule

	// orderSeq tracks the real committed partial order sequence number.
	orderSeq map[uint64]uint64

	//============================= internal interfaces =========================================

	// reader is used to read partial orders from meta pool tracker.
	reader internal.MetaReader

	//============================= metrics =================================

	// totalLogs tracks the number of committed partial order logs.
	totalLogs int

	//
	intervalLogs int

	// totalLatency tracks the total latency since partial order generation to commitment.
	totalLatency int64

	//
	intervalLatency int64

	//
	totalStreams int

	//
	intervalStreams int

	//
	totalCommitStreamLatency int64

	//
	intervalCommitStreamLatency int64

	//============================== external interfaces ==========================================

	// logger is used to print logs.
	logger external.Logger

	//
	streamMutex sync.Mutex

	//
	streams *list.List

	//
	count int64
}

func NewExecutor(oLeader, author uint64, n int, mgr internal.MetaPool, manager internal.TxManager, exec external.ExecutionService, logger external.Logger) *executorImpl {
	orderSeq := make(map[uint64]uint64)

	for i := 0; i < n; i++ {
		id := uint64(i + 1)
		orderSeq[id] = uint64(0)
	}

	cRecorder := recorder.NewCommandRecorder(author, n, logger)
	return &executorImpl{
		author:    author,
		rules:     newOrderRule(oLeader, author, n, cRecorder, mgr, mgr, manager, exec, logger),
		cRecorder: cRecorder,
		reader:    mgr,
		logger:    logger,
		orderSeq:  orderSeq,
		streams:   list.New(),
	}
}

// CommitStream is used to commit the partial order stream.
func (ei *executorImpl) CommitStream(qStream types.QueryStream) {
	ei.streamMutex.Lock()
	ei.streams.PushBack(qStream)
	ei.streamMutex.Unlock()
	atomic.AddInt64(&ei.count, 1)
}

func (ei *executorImpl) commitStream(qStream types.QueryStream) {
	if len(qStream) == 0 {
		// nil partial order batch means we should skip the current commitment attempt.
		return
	}

	start := time.Now()

	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	partials := ei.reader.ReadPartials(qStream)

	var oStream types.OrderStream

	for _, pOrder := range partials {
		sub := time.Now().UnixNano() - pOrder.OrderedTime

		// collect order log metrics.
		ei.totalLogs++
		ei.totalLatency += sub

		ei.intervalLogs++
		ei.intervalLatency += sub

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

	sub := time.Now().Sub(start).Milliseconds()
	ei.totalCommitStreamLatency += sub
	ei.totalStreams++
	ei.intervalCommitStreamLatency += sub
	ei.intervalStreams++
}

func (ei *executorImpl) Run() {
	for {
		if atomic.LoadInt64(&ei.count) == 0 {
			continue
		}

		ei.streamMutex.Lock()
		e := ei.streams.Front()
		atomic.AddInt64(&ei.count, -1)
		qStream := e.Value.(types.QueryStream)
		ei.streams.Remove(e)
		ei.streamMutex.Unlock()

		ei.commitStream(qStream)
	}
}

func (ei *executorImpl) Quit() {

}

// QueryMetrics returns metrics info of executor.
func (ei *executorImpl) QueryMetrics() types.MetricsInfo {
	return types.MetricsInfo{
		AveLogLatency:            ei.aveLogLatency(),
		CurLogLatency:            ei.curLogLatency(),
		AveCommandInfoLatency:    ei.aveCommandInfoLatency(),
		CurCommandInfoLatency:    ei.curCommandInfoLatency(),
		AveCommitStreamLatency:   ei.aveCommitStreamLatency(),
		CurCommitStreamLatency:   ei.curCommitStreamLatency(),
		SafeCommandCount:         ei.rules.totalSafeCommit,
		RiskCommandCount:         ei.rules.totalRiskCommit,
		FrontAttackFromRisk:      ei.rules.frontAttackFromRisk,
		FrontAttackFromSafe:      ei.rules.frontAttackFromSafe,
		FrontAttackIntervalRisk:  ei.rules.frontAttackIntervalRisk,
		FrontAttackIntervalSafe:  ei.rules.frontAttackIntervalSafe,
		MSafeCommandCount:        ei.rules.mediumCommit.totalSafeCommit,
		MRiskCommandCount:        ei.rules.mediumCommit.totalRiskCommit,
		MFrontAttackFromRisk:     ei.rules.mediumCommit.frontAttackFromRisk,
		MFrontAttackFromSafe:     ei.rules.mediumCommit.frontAttackFromSafe,
		MFrontAttackIntervalRisk: ei.rules.mediumCommit.frontAttackIntervalRisk,
		MFrontAttackIntervalSafe: ei.rules.mediumCommit.frontAttackIntervalSafe,
	}
}

// aveLogLatency returns average latency of partial orders to be committed.
func (ei *executorImpl) aveLogLatency() float64 {
	if ei.totalLogs == 0 {
		return 0
	}
	return types.NanoToMillisecond(ei.totalLatency / int64(ei.totalLogs))
}

// curLogLatency returns average latency of partial orders to be committed.
func (ei *executorImpl) curLogLatency() float64 {
	if ei.intervalLogs == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(ei.intervalLatency / int64(ei.intervalLogs))
	ei.intervalLatency = 0
	ei.intervalLogs = 0
	return ret
}

// aveCommandInfoLatency returns average latency of command info to be committed.
func (ei *executorImpl) aveCommandInfoLatency() float64 {
	if ei.rules.commit.totalCommandInfo == 0 {
		return 0
	}
	return types.NanoToMillisecond(ei.rules.commit.totalLatency / int64(ei.rules.commit.totalCommandInfo))
}

// curCommandInfoLatency returns average latency of command info to be committed.
func (ei *executorImpl) curCommandInfoLatency() float64 {
	if ei.rules.commit.intervalCommandInfo == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(ei.rules.commit.intervalLatency / int64(ei.rules.commit.intervalCommandInfo))
	ei.rules.commit.intervalLatency = 0
	ei.rules.commit.intervalCommandInfo = 0
	return ret
}

// aveCommitStreamLatency returns average latency of commitment of query stream.
func (ei *executorImpl) aveCommitStreamLatency() float64 {
	if ei.totalStreams == 0 {
		return 0
	}
	return types.NanoToMillisecond(ei.totalCommitStreamLatency / int64(ei.totalStreams))
}

// curCommitStreamLatency returns average latency of commitment of query stream.
func (ei *executorImpl) curCommitStreamLatency() float64 {
	if ei.intervalStreams == 0 {
		return 0
	}
	ret := types.NanoToMillisecond(ei.intervalCommitStreamLatency / int64(ei.intervalStreams))
	ei.intervalCommitStreamLatency = 0
	ei.intervalStreams = 0
	return ret
}
