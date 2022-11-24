package instance

import (
	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/google/btree"
	"sync"
)

// receive order-attempt and verify it.

type sequencerInstance struct {
	// mutex is used to resolve concurrency problems.
	mutex sync.Mutex

	//===================================== basic information =========================================

	// author is the local node's identifier.
	author uint64

	// isByzantine indicates if current node is untrusted.
	isByzantine bool

	// id indicates which node the current request-pool is maintained for.
	id uint64

	// cache is used to track the order-attempts.
	cache *btree.BTree

	//==================================== sub-chain management =============================================

	// highAttempt is used to track the latest order-attempt we have received.
	highAttempt *protos.OrderAttempt

	//======================================= internal modules =========================================

	// aTracker is the storage for order-attempt.
	aTracker api.AttemptTracker

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencerInstance(author, id uint64, aTracker api.AttemptTracker, logger external.Logger) api.SequencerInstance {
	return &sequencerInstance{
		author:   author,
		id:       id,
		cache:    btree.New(2),
		aTracker: aTracker,
		logger:   logger,
	}
}

func (si *sequencerInstance) GetHighAttempt() *protos.OrderAttempt {
	si.mutex.Lock()
	defer si.mutex.Unlock()

	if si.isByzantine {
		// untrusted node.
		return nil
	}

	return si.highAttempt
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

		si.highAttempt = min
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
	return attempt.SeqNo == si.highAttempt.SeqNo+1
}

func (si *sequencerInstance) verifyDigest(attempt *protos.OrderAttempt) bool {
	if attempt.ParentDigest != si.highAttempt.Digest {
		return false
	}
	return types.CheckOrderAttemptDigest(attempt)
}
