package execsimple

import (
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/recorder"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"sort"
	"sync"
	"time"
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

	// totalLatency tracks the total latency since partial order generation to commitment.
	totalLatency int64

	//============================== external interfaces ==========================================

	// logger is used to print logs.
	logger external.Logger
}

func NewExecutor(oLeader, author uint64, n int, mgr internal.MetaPool, manager internal.TxManager, exec external.ExecutionService, logger external.Logger) *executorImpl {
	orderSeq := make(map[uint64]uint64)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
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
	}
}

// CommitStream is used to commit the partial order stream.
func (ei *executorImpl) CommitStream(qStream types.QueryStream) error {
	ei.mutex.Lock()
	defer ei.mutex.Unlock()

	if len(qStream) == 0 {
		// nil partial order batch means we should skip the current commitment attempt.
		return nil
	}

	ei.logger.Debugf("[%d] commit query stream len %d: %v", ei.author, len(qStream), qStream)

	partials := ei.reader.ReadPartials(qStream)

	var oStream types.OrderStream

	for _, pOrder := range partials {
		// collect order log metrics.
		ei.totalLogs++
		ei.totalLatency += time.Now().UnixNano() - pOrder.OrderedTime

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

	if !updated {
		return nil
	}

	ei.rules.processPartialOrder()

	return nil
}

// QueryMetrics returns metrics info of executor.
func (ei *executorImpl) QueryMetrics() types.MetricsInfo {
	ei.mutex.RLock()
	defer ei.mutex.RUnlock()

	return types.MetricsInfo{
		AveLogLatency:           ei.aveLogLatency(),
		AveCommandInfoLatency:   ei.aveCommandInfoLatency(),
		SafeCommandCount:        ei.rules.totalSafeCommit,
		RiskCommandCount:        ei.rules.totalRiskCommit,
		FrontAttackFromRisk:     ei.rules.frontAttackFromRisk,
		FrontAttackFromSafe:     ei.rules.frontAttackFromSafe,
		FrontAttackIntervalRisk: ei.rules.frontAttackIntervalRisk,
		FrontAttackIntervalSafe: ei.rules.frontAttackIntervalSafe,
	}
}

// aveOrderLatency returns average latency of partial orders to be committed.
func (ei *executorImpl) aveLogLatency() float64 {
	if ei.totalLogs == 0 {
		return 0
	}
	return types.NanoToMillisecond(ei.totalLatency / int64(ei.totalLogs))
}

// aveCommandInfoLatency returns average latency of command info to be committed.
func (ei *executorImpl) aveCommandInfoLatency() float64 {
	if ei.rules.commit.totalCommandInfo == 0 {
		return 0
	}
	return types.NanoToMillisecond(ei.rules.commit.totalLatency / int64(ei.rules.commit.totalCommandInfo))
}
