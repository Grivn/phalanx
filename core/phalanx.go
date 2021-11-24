package phalanx

import (
	"fmt"
	"github.com/Grivn/phalanx/executor/execsimple"

	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"
	"github.com/Grivn/phalanx/metapool"
	"github.com/Grivn/phalanx/txmanager"

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
}

func NewPhalanxProvider(n int, author uint64, commandSize int, exec external.ExecutionService, network external.NetworkService, logger external.Logger) *phalanxImpl {
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
	txMgr := txmanager.NewTxManager(n, author, commandSize, network, mLogs.txManagerLog)

	// initiate meta pool.
	mPool := metapool.NewMetaPool(n, author, network, mLogs.metaPoolLog)

	// initiate executor.
	executor := execsimple.NewExecutor(author, n, mPool, exec, mLogs.executorLog)

	return &phalanxImpl{
		author:    author,
		txManager: txMgr,
		metaPool:  mPool,
		executor:  executor,
		logger:    logger,
	}
}

func (phi *phalanxImpl) Run() {
	go phi.metaPool.Run()
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

	return phi.executor.CommitStream(qStream)
}
