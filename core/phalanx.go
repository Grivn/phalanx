package phalanx

import (
	"fmt"
	"github.com/Grivn/phalanx/executor/order"
	"time"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/metapool"
	"github.com/Grivn/phalanx/receiver"
	"github.com/gogo/protobuf/proto"
)

type phalanxImpl struct {
	// author indicates the identifier of current node.
	author uint64

	// txManager is used to process transactions.
	txManager internal.TxManager

	// metaPool is used to process meta consensus data:
	// 1) partial consensus for order logs.
	// 2) global consensus for total order.
	// 3) cache essential data information.
	metaPool internal.MetaPool

	// executor is used to generate the final ordered blocks.
	executor internal.Executor

	// logger is used to print logs.
	logger external.Logger

	//
	logCount int
}

func NewPhalanxProvider(oLeader uint64, byz bool, openLatency int, duration time.Duration, interval int, cDuration time.Duration, n int, multi int, logCount int, memSize int, author uint64, commandSize int, exec external.ExecutionService, network external.NetworkService, logger external.Logger, selected uint64) *phalanxImpl {
	// todo read crypto key pairs from config files.
	// initiate key pairs.
	_ = crypto.SetKeys()

	// initiate phalanx logger.
	mLogs, err := newPLogger(logger, true, author)
	if err != nil {
		logger.Errorf("Generate Phalanx Logger Failed: %s", err)
		return nil
	}

	// initiate tx manager.
	txMgr := receiver.NewTxManager(multi, author, commandSize, memSize, network, mLogs.txManagerLog, selected)

	// initiate meta pool.
	mPool := metapool.NewMetaPool(byz, openLatency, duration, n, multi, logCount, author, network, mLogs.metaPoolLog)

	// initiate executor.
	executor := order.NewExecutor(oLeader, author, n, mPool, txMgr, exec, mLogs.executorLog)

	return &phalanxImpl{
		author:    author,
		txManager: txMgr,
		metaPool:  mPool,
		executor:  executor,
		logger:    logger,
		logCount:  logCount,
	}
}

func (phi *phalanxImpl) Run() {
	go phi.metaPool.Run()
	go phi.txManager.Run()
	go phi.executor.Run()
}

func (phi *phalanxImpl) Quit() {
	phi.metaPool.Quit()
}

// ReceiveTransaction is used to process transaction we have received.
func (phi *phalanxImpl) ReceiveTransaction(tx *protos.Transaction) {
	phi.txManager.ProcessTransaction(tx)
}

// ReceiveCommand is used to process the commands from clients.
func (phi *phalanxImpl) ReceiveCommand(command *protos.Command) {
	phi.metaPool.ProcessCommand(command)
}

// ReceiveConsensusMessage is used process the consensus messages from phalanx replica.
func (phi *phalanxImpl) ReceiveConsensusMessage(message *protos.ConsensusMessage) error {
	switch message.Type {
	case protos.MessageType_PRE_ORDER:
		pre := &protos.PreOrder{}
		if err := proto.Unmarshal(message.Payload, pre); err != nil {
			return fmt.Errorf("unmarshal error: %s", err)
		}
		if err := phi.metaPool.ProcessPreOrder(pre); err != nil {
			phi.logger.Errorf("[%d] failed process pre-order, error msg: %s", phi.author, err)
		}
	case protos.MessageType_QUORUM_CERT:
		pOrder := &protos.PartialOrder{}
		if err := proto.Unmarshal(message.Payload, pOrder); err != nil {
			return fmt.Errorf("unmarshal error: %s", err)
		}
		if err := phi.metaPool.ProcessPartial(pOrder); err != nil {
			phi.logger.Errorf("[%d] failed process partial-order, error msg: %s", phi.author, err)
		}
	case protos.MessageType_VOTE:
		vote := &protos.Vote{}
		if err := proto.Unmarshal(message.Payload, vote); err != nil {
			return fmt.Errorf("unmarshal error: %s", err)
		}
		if err := phi.metaPool.ProcessVote(vote); err != nil {
			phi.logger.Errorf("[%d] failed process vote, error msg: %s", phi.author, err)
		}
	}
	return nil
}

// MakeProposal is used to generate phalanx proposal for consensus.
func (phi *phalanxImpl) MakeProposal() (*protos.PartialOrderBatch, error) {
	return phi.metaPool.GenerateProposal()
}

// CommitProposal is used to commit the phalanx proposal which has been verified with consensus.
func (phi *phalanxImpl) CommitProposal(pBatch *protos.PartialOrderBatch) error {
	phi.logger.Debugf("[%d] received partial batch %s", phi.author, pBatch.Format())

	qStream, err := phi.metaPool.VerifyProposal(pBatch)
	if err != nil {
		phi.logger.Errorf("verify error, %s", err)
		return err
	}
	phi.executor.CommitStream(qStream)
	return nil
}

// QueryMetrics returns the metrics info of phalanx.
func (phi *phalanxImpl) QueryMetrics() types.MetricsInfo {
	execMetrics := phi.executor.QueryMetrics()
	metaMetrics := phi.metaPool.QueryMetrics()
	return types.MetricsInfo{
		AveOrderSize:             metaMetrics.AveOrderSize,
		AvePackOrderLatency:      metaMetrics.AvePackOrderLatency,
		CurPackOrderLatency:      metaMetrics.CurPackOrderLatency,
		AveOrderLatency:          metaMetrics.AveOrderLatency,
		CurOrderLatency:          metaMetrics.CurOrderLatency,
		AveLogLatency:            execMetrics.AveLogLatency,
		CurLogLatency:            execMetrics.CurLogLatency,
		AveCommandInfoLatency:    execMetrics.AveCommandInfoLatency,
		CurCommandInfoLatency:    execMetrics.CurCommandInfoLatency,
		AveCommitStreamLatency:   execMetrics.AveCommitStreamLatency,
		CurCommitStreamLatency:   execMetrics.CurCommitStreamLatency,
		SafeCommandCount:         execMetrics.SafeCommandCount,
		RiskCommandCount:         execMetrics.RiskCommandCount,
		FrontAttackFromRisk:      execMetrics.FrontAttackFromRisk,
		FrontAttackFromSafe:      execMetrics.FrontAttackFromSafe,
		FrontAttackIntervalRisk:  execMetrics.FrontAttackIntervalRisk,
		FrontAttackIntervalSafe:  execMetrics.FrontAttackIntervalSafe,
		CommandPS:                metaMetrics.CommandPS,
		LogPS:                    metaMetrics.LogPS,
		GenLogPS:                 metaMetrics.GenLogPS,
		MSafeCommandCount:        execMetrics.MSafeCommandCount,
		MRiskCommandCount:        execMetrics.MRiskCommandCount,
		MFrontAttackFromRisk:     execMetrics.MFrontAttackFromRisk,
		MFrontAttackFromSafe:     execMetrics.MFrontAttackFromSafe,
		MFrontAttackIntervalRisk: execMetrics.MFrontAttackIntervalRisk,
		MFrontAttackIntervalSafe: execMetrics.MFrontAttackIntervalSafe,
	}
}
