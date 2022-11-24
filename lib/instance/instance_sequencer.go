package instance

import (
	"fmt"
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/google/btree"
	"sync"
)

type sequencerInstance struct {
	// mutex is used to resolve concurrency problems.
	mutex sync.Mutex

	isByzantine bool

	n int

	quorum int

	//===================================== basic information =========================================

	// author is the local node's identifier.
	author uint64

	// id indicates which node the current request-pool is maintained for.
	id uint64

	// sequence is the preferred seqNo for the next request.
	sequence uint64

	// cache is used to track the order-attempts.
	cache *btree.BTree

	//==================================== sub-chain management =============================================

	// latestAttempt is used to track the latest order-attempt we have received.
	latestAttempt *protos.OrderAttempt

	// minimumInterval indicates the minimum number of order attempts to create checkpoint each time.
	minimumInterval uint64

	// latestCheckpoint is used to track the latest checkpoint we have received.
	latestCheckpoint *protos.Checkpoint

	// aggMap is used to make aggregation quorum certificate for checkpoints.
	aggMap map[string]*CheckpointEvent

	//======================================= internal modules =========================================

	// aTracker is the storage for order-attempt.
	aTracker api.AttemptTracker

	crypto api.Crypto

	//======================================= external tools ===========================================

	eventC chan interface{}

	closeC chan bool

	outputC chan *protos.Checkpoint

	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}

func (si *sequencerInstance) Append(attempt *protos.OrderAttempt) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	si.cache.ReplaceOrInsert(attempt)

	min := si.minAttempt()
	for {
		if min == nil {
			break
		}

		si.latestAttempt = min
		si.aTracker.Record(min)
		break
	}
}

func (si *sequencerInstance) minAttempt() *protos.OrderAttempt {
	item := si.cache.Min()
	if item == nil {
		return nil
	}

	attempt, ok := item.(*protos.OrderAttempt)
	if !ok {
		return nil
	}

	if !si.verifySeqNo(attempt) {
		return nil
	}

	if !si.verifyDigest(attempt) {
		si.isByzantine = true
		return nil
	}

	return attempt
}

func (si *sequencerInstance) verifySeqNo(attempt *protos.OrderAttempt) bool {
	return attempt.SeqNo == si.latestAttempt.SeqNo+1
}

func (si *sequencerInstance) verifyDigest(attempt *protos.OrderAttempt) bool {
	if attempt.ParentDigest != si.latestAttempt.Digest {
		return false
	}
	return types.CheckOrderAttemptDigest(attempt)
}

func (si *sequencerInstance) GetLatestCheckpoint() *protos.Checkpoint {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	if si.isByzantine {
		// untrusted node.
		return nil
	}

	if si.latestCheckpoint != nil && si.latestCheckpoint.OrderAttempt.SeqNo-si.latestAttempt.SeqNo <= si.minimumInterval {
		return si.latestCheckpoint
	}

	return nil
}

func (si *sequencerInstance) RequestCheckpoint() (*CheckpointEvent, error) {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	if si.isByzantine {
		return nil, nil
	}

	if si.latestAttempt == nil {
		return nil, nil
	}

	attempt := si.latestAttempt
	if event, ok := si.aggMap[attempt.Digest]; ok {
		return event, nil
	}

	event := &CheckpointEvent{
		Attempt:   si.latestAttempt,
		QC:        protos.NewQuorumCert(),
		CallbackC: make(chan *protos.Checkpoint),
	}

	signature, err := si.crypto.PrivateSign(types.StringToBytes(attempt.Digest))
	if err != nil {
		return nil, fmt.Errorf("generate signature for pre-order failed: %s", err)
	}
	event.QC.Certs[si.author] = signature

	request := protos.NewCheckpointRequest(si.author, attempt)
	cm, err := protos.PackCheckpointRequest(request)
	if err != nil {
		return nil, fmt.Errorf("generate consensus message error: %s", err)
	}
	si.sender.BroadcastPCM(cm)
	si.aggMap[attempt.Digest] = event

	return event, nil
}

func (si *sequencerInstance) ProcessCheckpointRequest(request *protos.CheckpointRequest) error {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	vote := protos.NewCheckpointVote(si.author, request)
	signature, err := si.crypto.PrivateSign(types.StringToBytes(request.OrderAttempt.Digest))
	if err != nil {
		return fmt.Errorf("generate signature for pre-order failed: %s", err)
	}
	vote.Cert = signature
	cm, err := protos.PackCheckpointVote(vote, request.Author)
	if err != nil {
		return fmt.Errorf("generate consensus message error: %s", err)
	}
	si.sender.BroadcastPCM(cm)
	return nil
}

func (si *sequencerInstance) ProcessCheckpointVote(vote *protos.CheckpointVote) error {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	event, ok := si.aggMap[vote.Digest]
	if !ok {
		// there are 2 conditions that a pre-order waiting for agg-sig cannot be found: 1) we have never generated such
		// a pre-order message, 2) we have already generated an order message for it, which means it has been verified.
		return nil
	}

	// verify the signature in vote
	// here, we would like to check if the signature is valid.
	if err := si.crypto.PublicVerify(vote.Cert, types.StringToBytes(vote.Digest), vote.Author); err != nil {
		return fmt.Errorf("failed to aggregate: %s", err)
	}

	// record the certification in current vote
	event.QC.Certs[vote.Author] = vote.Cert

	// check the quorum size for proof-certs
	if len(event.QC.Certs) == si.quorum {
		si.logger.Debugf("[%d] found quorum votes, generate quorum order %s", si.author, event.Attempt.Format())
		delete(si.aggMap, vote.Digest)
		checkpoint := event.ToCheckpoint()
		si.aTracker.Checkpoint(checkpoint)

		if checkpoint.OrderAttempt.SeqNo > si.latestCheckpoint.OrderAttempt.SeqNo {
			si.latestCheckpoint = checkpoint
		}

		go func() {
			event.CallbackC <- checkpoint
		}()
	}
	return nil
}
