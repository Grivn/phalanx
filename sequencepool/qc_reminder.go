package sequencepool

import (
	"errors"
	"fmt"
	"github.com/Grivn/phalanx/common/crypto"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/google/btree"
)

type qcReminder struct {
	// author indicates the identifier for current participate.
	author uint64

	// id is the identifier for current replica.
	id uint64

	// quorum indicates the legal size for bft.
	quorum int

	// cachedQCs is a cache for QCs that are generated by replicas and have not been proposed yet.
	cachedQCs *btree.BTree

	// proposedQCs is used to record the QCs which have been proposed.
	proposedQCs *btree.BTree

	// proposedNo is used to track the seqNo which has been proposed.
	proposedNo map[uint64]bool

	// stableNo indicates the latest stable seqNo.
	// stable sequence number: the number which has been verified by bft consensus module.
	stableNo uint64

	// seqNo is used for leader in synchronous bft to track the block generation, it is the next seqNo to propose.
	seqNo uint64
}

func newQCReminder(author uint64, n int, id uint64) *qcReminder {
	return &qcReminder{
		author:      author,
		id:          id,
		quorum:      types.CalculateQuorum(n),
		cachedQCs:   btree.New(2),
		proposedQCs: btree.New(2),
		proposedNo:  make(map[uint64]bool),
		seqNo:       uint64(1),
	}
}

// becomeLeader is used to init the replica which has just become the leader of cluster.
func (qr *qcReminder) becomeLeader() {
	qr.seqNo = qr.stableNo+1
}

// restoreQCs is used to restore the QCs from proposedQCs and remove these seqNo from proposedNo.
func (qr *qcReminder) restoreQCs() {
	// here, we would like to restore the QCs which have a larger seqNo than the stable one in proposedQCs, which means
	// these restored QCs should be re-proposed for block generation.
	for {
		// get the minQC from proposedQCs.
		minQC := qr.proposedDeleteMin()

		// there aren't any QCs in proposedQCs, stop the restore process.
		if minQC == nil {
			break
		}

		// the QCs with a expired seqNo do not need to be re-proposed for they have already been verified to generate a block.
		if minQC.Sequence() <= qr.stableNo {
			continue
		}

		// restore the QC which has not been verified yet.
		qr.cachedQCs.ReplaceOrInsert(minQC)

		// update the proposedNo map at the same time.
		delete(qr.proposedNo, minQC.Sequence())
	}
}

// insertQC is used to store the QCs generated by current replica, we would like to store the QCs by btree,
// so that we could easily find a minQC, and the QCs here should restrict the following rules:
// 1) stability: it should be larger than the stable seqNo.
// 2) availability: it should be a QC which has never been proposed yet.
func (qr *qcReminder) insertQC(qc *protos.QuorumCert) error {
	if qc.Sequence() <= qr.stableNo {
		return fmt.Errorf("expired seqNo: stable seqNo %d, received seqNo %d", qr.stableNo, qc.Sequence())
	}

	if qr.isProposed(qc) {
		return fmt.Errorf("proposed seqNo: received seqNo %d", qc.Sequence())
	}

	qr.cachedQCs.ReplaceOrInsert(qc)
	return nil
}

// pullQC is used to pull the QCs to generate the payload for bft.
func (qr *qcReminder) pullQC() *protos.QuorumCert {
	minQC := qr.cacheDeleteMin()

	if minQC == nil {
		return nil
	}

	if minQC.Sequence() != qr.seqNo {
		return nil
	}

	qr.seqNo++

	return minQC
}

// backQC is used to push back the QC and update the seqNo for payload generation.
func (qr *qcReminder) backQC(qc *protos.QuorumCert) {
	if qc.Sequence() < qr.seqNo {
		qr.seqNo = qc.Sequence()
	}

	qr.cachedQCs.ReplaceOrInsert(qc)
}

// setStableQC is used to make the verified QCs stable, in which we would like to update the stable seqNo and remove the expired QCs.
func (qr *qcReminder) setStableQC(qc *protos.QuorumCert) error {
	// check the validation of sequence number of stableQC
	if qc.Sequence() != qr.stableNo+1 {
		return fmt.Errorf("invalid seqNo: expected seqNo %d, received seqNo %d", qr.stableNo+1, qc.Sequence())
	}

	// update the stable seqNo.
	qr.stableNo = qc.Sequence()

	// clear the expired cachedQCs
	for {
		minQC := qr.cachedMin()

		if minQC == nil {
			break
		}

		if minQC.Sequence() > qr.stableNo {
			break
		}

		qr.cachedQCs.Delete(minQC)
	}

	// clear the expired proposedQCs
	for {
		minQC := qr.proposedMin()

		if minQC == nil {
			break
		}

		if minQC.Sequence() > qr.stableNo {
			break
		}

		qr.proposedQCs.Delete(minQC)
	}

	return nil
}

// verify is used to check the QC from remote participates when we are running a partial-synchronized byzantine consensus.
func (qr *qcReminder) verify(author uint64, remoteQC *protos.QuorumCert) error {
	// todo we need initial data for correct stable-sequence

	// the QCBatch is generated by self, skip verification.
	if author == qr.author {
		return nil
	}

	// the remote QC shouldn't be nil.
	if remoteQC == nil {
		return errors.New("nil remote QC")
	}

	// the No of remoteQC should be sequentially increased.
	if qr.proposedNo[remoteQC.Sequence()] {
		return fmt.Errorf("proposed seqNo: reminder-ID %d, received seqNo %d", qr.id, remoteQC.Sequence())
	}

	// the signature of QC should be valid.
	if err := crypto.VerifyProofCerts(types.StringToBytes(remoteQC.Digest()), remoteQC.ProofCerts, qr.quorum); err != nil {
		return fmt.Errorf("invalid QC signature: %s", err)
	}

	// remove the remoteQC from cache.
	qr.cachedDelete(remoteQC)

	// record the remoteQC as a proposed one.
	qr.proposedQC(remoteQC)

	return nil
}

// cachedMin is used to get the smallest QC from cache.
func (qr *qcReminder) cachedMin() *protos.QuorumCert {
	item := qr.cachedQCs.Min()

	if item == nil {
		return nil
	}

	return item.(*protos.QuorumCert)
}

// cacheDeleteMin is used to get the smallest QC from cache and delete it.
func (qr *qcReminder) cacheDeleteMin() *protos.QuorumCert {
	item := qr.cachedQCs.DeleteMin()

	if item == nil {
		return nil
	}

	return item.(*protos.QuorumCert)
}

// cachedDelete is used to delete the QC in cache.
func (qr *qcReminder) cachedDelete(qc *protos.QuorumCert) {
	qr.cachedQCs.Delete(qc)
}

func (qr *qcReminder) isProposed(qc *protos.QuorumCert) bool {
	return qr.proposedNo[qc.Sequence()]
}

// insertQC is used to store the QCs generated by current participate.
func (qr *qcReminder) proposedQC(qc *protos.QuorumCert) {
	// update the proposedNo
	qr.proposedNo[qc.Sequence()] = true

	// record the QC into proposedQCs map
	qr.proposedQCs.ReplaceOrInsert(qc)
}

// proposedMin is used to get the smallest QC from proposed.
func (qr *qcReminder) proposedMin() *protos.QuorumCert {
	item := qr.proposedQCs.Min()

	if item == nil {
		return nil
	}

	return item.(*protos.QuorumCert)
}

func (qr *qcReminder) proposedDeleteMin() *protos.QuorumCert {
	item := qr.proposedQCs.DeleteMin()

	if item == nil {
		return nil
	}

	return item.(*protos.QuorumCert)
}
