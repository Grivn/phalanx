package requester

import (
	commonProto "github.com/Grivn/phalanx/common/protos"
	commonTypes "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
)

type requestPool struct {
	// author is the local node's identifier
	author uint64

	// id indicates which node the current request-pool is maintained for
	id uint64

	// sequence is the preferred number for the next request
	sequence uint64

	// recorder is used to track the batch digest according to node identifier
	recorder map[uint64]string

	// recvC is the channel group which is used to receive events from other modules
	recvC commonTypes.RequesterRecvChan

	// sendC is the channel group which is used to send back information to other modules
	sendC commonTypes.RequesterSendChan

	// closeC is used to close the go-routine of request-pool
	closeC chan bool

	// logger is used to print logs
	logger external.Logger
}

func newRequestPool(author, id uint64, sendC commonTypes.RequesterSendChan, logger external.Logger) *requestPool {
	logger.Noticef("replica %d init request pool for replica %d", author, id)

	recvC := commonTypes.RequesterRecvChan{
		ProposalChan: make(chan *commonProto.Proposal),
	}

	return &requestPool{
		author:   author,
		id:       id,
		sequence: uint64(0),
		recorder: make(map[uint64]string),
		recvC:    recvC,
		sendC:    sendC,
		closeC:   make(chan bool),
		logger:   logger,
	}
}

func (rp *requestPool) start() {
	go rp.listener()
}

func (rp *requestPool) stop() {
	select {
	case <-rp.closeC:
	default:
		close(rp.closeC)
	}
}

func (rp *requestPool) record(proposal *commonProto.Proposal) {
	rp.recvC.ProposalChan <- proposal
}

func (rp *requestPool) listener() {
	for {
		select {
		case <-rp.closeC:
			rp.logger.Noticef("exist requestRecorderMgr listener for %d", rp.id)
			return
		case proposal, ok := <-rp.recvC.ProposalChan:
			if !ok {
				continue
			}
			rp.processProposal(proposal)
		}
	}
}

func (rp *requestPool) processProposal(proposal *commonProto.Proposal) {
	rp.logger.Infof("replica %d receive proposal %s", rp.author, proposal.Format())

	if _, ok := rp.recorder[proposal.Sequence]; ok {
		rp.logger.Warningf("replica %d has already received batch for replica %d sequence %d", rp.author, proposal.Author, proposal.Sequence)
		return
	}

	rp.recorder[proposal.Sequence] = proposal.TxBatch.Digest
	for {
		bid, ok := rp.recorder[rp.sequence+1]
		if !ok {
			break
		}
		rp.sequence++
		rp.logger.Infof("replica %d propose batch id for replica %d sequence %d", rp.author, rp.id, rp.sequence)

		go rp.sendSequentialBatch(bid)
		delete(rp.recorder, rp.sequence)
	}
}

func (rp *requestPool) sendSequentialBatch(bid *commonProto.BatchId) {
	rp.sendC.BatchIdChan <- bid
}
