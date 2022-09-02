package phalanx

import (
	"fmt"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/executor/finality"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metapool"
	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/receiver"
	"github.com/gogo/protobuf/proto"
)

type phalanxImpl struct {
	// author indicates the identifier of current node.
	author uint64

	// txManager is used to process transactions.
	txManager api.TxManager

	// metaPool is used to process meta consensus data:
	// 1) partial consensus for order logs.
	// 2) global consensus for total order.
	// 3) cache essential data information.
	metaPool api.MetaPool

	// executor is used to generate the final ordered blocks.
	executor api.Finality

	//
	metrics *metrics.Metrics

	// logger is used to print logs.
	logger external.Logger

	//
	logCount int
}

func NewPhalanxProvider(conf Config) *phalanxImpl {
	// todo read crypto key pairs from config files.
	// initiate key pairs.
	_ = crypto.SetKeys()

	// initiate phalanx logger.
	mLogs, err := newPLogger(conf.Logger, true, conf.Author)
	if err != nil {
		conf.Logger.Errorf("Generate Phalanx Logger Failed: %s", err)
		return nil
	}

	// create metrics.
	pMetrics := metrics.NewMetrics()

	// initiate tx manager.
	txConf := receiver.Config{
		Author:      conf.Author,
		Multi:       conf.Multi,
		CommandSize: conf.CommandSize,
		MemSize:     conf.MemSize,
		Selected:    conf.Selected,
		Sender:      conf.Network,
		Logger:      mLogs.txManagerLog,
	}
	txMgr := receiver.NewTxManager(txConf)

	// initiate meta pool.
	mpConf := metapool.Config{
		Author:   conf.Author,
		Byz:      conf.Byz,
		N:        conf.N,
		Multi:    conf.Multi,
		LogCount: conf.LogCount,
		Sender:   conf.Network,
		Logger:   mLogs.metaPoolLog,
		Metrics:  pMetrics.MetaPoolMetrics,
	}
	mPool := metapool.NewMetaPool(mpConf)

	// initiate executor.
	exeConf := finality.Config{
		Author:  conf.Author,
		OLeader: conf.OLeader,
		N:       conf.N,
		Mgr:     mPool,
		Manager: txMgr,
		Exec:    conf.Exec,
		Logger:  mLogs.executorLog,
		Metrics: pMetrics,
	}
	executor := finality.NewFinality(exeConf)

	return &phalanxImpl{
		author:    conf.Author,
		txManager: txMgr,
		metaPool:  mPool,
		executor:  executor,
		logger:    conf.Logger,
		logCount:  conf.LogCount,
		metrics:   pMetrics,
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
	return phi.metrics.QueryMetrics()
}
