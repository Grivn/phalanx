package phalanx

import (
	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/service/conengine"
	"github.com/Grivn/phalanx/pkg/service/crypto"
	"github.com/Grivn/phalanx/pkg/service/finality"
	"github.com/Grivn/phalanx/pkg/service/finengine"
	"github.com/Grivn/phalanx/pkg/service/memorypool"
	"github.com/Grivn/phalanx/pkg/service/receiver"
	"github.com/Grivn/phalanx/pkg/service/seqengine"
	"github.com/Grivn/phalanx/pkg/service/sequencer"
	"github.com/Grivn/phalanx/pkg/utils/tracker"
	"github.com/gogo/protobuf/proto"
)

type phalanxV2Impl struct {
	author           uint64
	proposer         api.Proposer
	sequencer        api.Sequencer
	sequencingEngine api.SequencingEngine
	memoryPool       api.MemoryPool
	consensusEngine  api.ConsensusEngine
	finality         api.Finality
	finalityEngine   api.FinalityEngine
	metrics          *metrics.Metrics
	logger           external.Logger
}

func NewPhalanxV2(
	conf config.PhalanxConf,
	privateKey external.PrivateKey,
	publicKeys external.PublicKeys,
	executor external.Executor,
	sender external.Sender,
	logger external.Logger) ProviderV2 {

	// initiate phalanx logger.
	mLogs, err := newPLoggerV2(logger, true, conf.NodeID)
	if err != nil {
		logger.Errorf("Generate Phalanx Logger Failed: %s", err)
		return nil
	}

	prop := receiver.NewTxManager(conf, sender, mLogs.proposerLog)
	commandTracker := tracker.NewCommandTracker(conf.NodeID, logger)
	attemptTracker := tracker.NewAttemptTracker(conf.NodeID, logger)
	checkpointTracker := tracker.NewCheckpointTracker(conf.NodeID, logger)
	cryptoService := crypto.NewCryptoService(privateKey, publicKeys)
	sequencingEngine := seqengine.NewSequencingEngine(conf, mLogs.sequencingEngineLog)
	seq := sequencer.NewSequencer(conf, sequencingEngine, sender, mLogs.sequencerLog)
	consensusEngine := conengine.NewConsensusEngine(conf, checkpointTracker, cryptoService, sender, mLogs.consensusEngineLog)
	mem := memorypool.NewMemoryPool(conf, consensusEngine, commandTracker, attemptTracker, checkpointTracker, cryptoService, mLogs.memoryPoolLog)
	finalityEngine := finengine.NewPhalanxAnchorBasedOrdering(conf, commandTracker, executor, mLogs.finalityEngine)
	fin := finality.NewFinality(conf, attemptTracker, finalityEngine, mLogs.finalityLog)

	return &phalanxV2Impl{
		author:           conf.NodeID,
		proposer:         prop,
		sequencer:        seq,
		sequencingEngine: sequencingEngine,
		memoryPool:       mem,
		consensusEngine:  consensusEngine,
		finality:         fin,
		finalityEngine:   finalityEngine,
		metrics:          metrics.NewMetrics(),
		logger:           logger,
	}
}

func (p *phalanxV2Impl) Run() {
	p.proposer.Run()
	p.consensusEngine.Run()
	p.sequencer.Run()
	p.finality.Run()
}

func (p *phalanxV2Impl) Quit() {
	p.proposer.Quit()
	p.consensusEngine.Quit()
	p.sequencer.Quit()
	p.finality.Quit()
}

// ReceiveTransaction is used to process transaction we have received.
func (p *phalanxV2Impl) ReceiveTransaction(tx *protos.Transaction) {
	p.proposer.ProcessTransaction(tx)
}

// ReceiveCommand is used to process the commands from clients.
func (p *phalanxV2Impl) ReceiveCommand(command *protos.Command) {
	p.sequencer.Sequencing(command)
	p.memoryPool.ProcessCommand(command)
}

func (p *phalanxV2Impl) ReceiveOrderAttempt(attempt *protos.OrderAttempt) {
	p.memoryPool.ProcessOrderAttempt(attempt)
}

// ReceiveConsensusMessage is used process the consensus messages from phalanx replica.
func (p *phalanxV2Impl) ReceiveConsensusMessage(message *protos.ConsensusMessage) {
	if message.Type == protos.MessageType_ORDER_ATTEMPT {
		attemtp := &protos.OrderAttempt{}
		proto.Unmarshal(message.Payload, attemtp)
		go p.memoryPool.ProcessOrderAttempt(attemtp)
		return
	}
	go p.consensusEngine.ProcessConsensusMessage(message)
}

// MakeProposal is used to generate phalanx proposal for consensus.
func (p *phalanxV2Impl) MakeProposal() (*protos.Proposal, error) {
	return p.memoryPool.GenerateProposal()
}

// CommitProposal is used to commit the phalanx proposal which has been verified with consensus.
func (p *phalanxV2Impl) CommitProposal(pBatch *protos.Proposal) error {
	p.logger.Debugf("[%d] received partial batch %s", p.author, pBatch.Format())

	qStream, err := p.memoryPool.VerifyProposal(pBatch)
	if err != nil {
		p.logger.Errorf("verify error, %s", err)
		return err
	}
	p.finality.CommitStream(qStream)
	return nil
}

// QueryMetrics returns the metrics info of phalanx.
func (p *phalanxV2Impl) QueryMetrics() types.MetricsInfo {
	return p.metrics.QueryMetrics()
}
