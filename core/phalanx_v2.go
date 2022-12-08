package phalanx

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/metrics"
)

type phalanxV2Impl struct {
	author           uint64
	proposer         api.Proposer
	sequencer        api.Sequencer
	sequencingEngine api.SequencingEngine
	memoryPool       api.MemoryPool
	consensusEngine  api.ConsensusEngine
	finality         api.Finality
	metrics          *metrics.Metrics
	logger           external.Logger
}

func (p *phalanxV2Impl) Run() {
	go p.proposer.Run()
	go p.consensusEngine.Run()
	go p.sequencer.Run()
	go p.finality.Run()
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
	p.memoryPool.ProcessCommand(command)
}

func (p *phalanxV2Impl) ReceiveOrderAttempt(attempt *protos.OrderAttempt) {
	p.memoryPool.ProcessOrderAttempt(attempt)
}

// ReceiveConsensusMessage is used process the consensus messages from phalanx replica.
func (p *phalanxV2Impl) ReceiveConsensusMessage(message *protos.ConsensusMessage) {
	p.consensusEngine.ProcessConsensusMessage(message)
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
