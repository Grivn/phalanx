package instance

import (
	"sync"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/google/btree"
)

// receive order-attempt and verify it.

type sequencerInstance struct {
	// mutex is used to resolve concurrency problems.
	mutex sync.RWMutex

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

	// highestAttempt is used to track the latest order-attempt we have received.
	highestAttempt *protos.OrderAttempt

	//======================================= internal modules =========================================

	// logger is used to print logs.
	logger external.Logger
}

func NewSequencerInstance(author, id uint64, logger external.Logger) api.SequencerInstance {
	return &sequencerInstance{
		author: author,
		id:     id,
		cache:  btree.New(2),
		logger: logger,
	}
}

func (si *sequencerInstance) GetHighestAttempt() *protos.OrderAttempt {
	si.mutex.RLock()
	defer si.mutex.RUnlock()

	if si.isByzantine {
		// untrusted node.
		return nil
	}

	return si.highestAttempt
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

		si.highestAttempt = min
		min = si.minAttempt()
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
	return attempt.SeqNo == si.highestAttempt.SeqNo+1
}

func (si *sequencerInstance) verifyDigest(attempt *protos.OrderAttempt) bool {
	if attempt.ParentDigest != si.highestAttempt.Digest {
		return false
	}
	return types.CheckOrderAttemptDigest(attempt)
}
