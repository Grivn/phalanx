package phalanx

import (
	"fmt"

	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/service/finality_v1"
	"github.com/Grivn/phalanx/pkg/service/metapool"
	"github.com/Grivn/phalanx/pkg/service/receiver"
	"github.com/gogo/protobuf/proto"
)

type phalanxImpl struct {
	// author indicates the identifier of current node.
	author uint64

	// proposer is used to process transactions.
	proposer api.Proposer

	// metaPool is used to process meta consensus data:
	// 1) partial consensus for order logs.
	// 2) global consensus for total order.
	// 3) cache essential data information.
	metaPool api.MetaPool

	// finality is used to generate the final ordered blocks.
	finality api.Finality

	// metrics is used to record the metric of current phalanx instance.
	metrics *metrics.Metrics

	// logger is used to print logs.
	logger external.Logger
}

func NewPhalanxProvider(
	conf config.PhalanxConf,
	privateKey external.PrivateKey,
	publicKeys external.PublicKeys,
	executor external.Executor,
	sender external.Sender,
	logger external.Logger) *phalanxImpl {
	// todo read crypto key pairs from config files.
	// initiate key pairs.

	// initiate phalanx logger.
	mLogs, err := newPLogger(logger, true, conf.NodeID)
	if err != nil {
		logger.Errorf("Generate Phalanx Logger Failed: %s", err)
		return nil
	}

	// create metrics.
	ms := metrics.NewMetrics()

	// initiate tx manager.
	proposer := receiver.NewTxManager(conf, sender, mLogs.txManagerLog)

	// initiate meta pool.
	mPool := metapool.NewMetaPool(conf, privateKey, publicKeys, sender, mLogs.metaPoolLog, ms)

	// initiate executor.
	final := finality_v1.NewFinality(conf, mPool, executor, mLogs.executorLog, ms)

	return &phalanxImpl{
		author:   conf.NodeID,
		proposer: proposer,
		metaPool: mPool,
		finality: final,
		logger:   logger,
		metrics:  ms,
	}
}

func (phi *phalanxImpl) Run() {
	phi.metaPool.Run()
	phi.proposer.Run()
	phi.finality.Run()
}

func (phi *phalanxImpl) Quit() {
	phi.metaPool.Quit()
}

// ReceiveTransaction is used to process transaction we have received.
func (phi *phalanxImpl) ReceiveTransaction(tx *protos.Transaction) {
	phi.proposer.ProcessTransaction(tx)
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
	phi.finality.CommitStream(qStream)
	return nil
}

// QueryMetrics returns the metrics info of phalanx.
func (phi *phalanxImpl) QueryMetrics() types.MetricsInfo {
	return phi.metrics.QueryMetrics()
}
