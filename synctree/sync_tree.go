package synctree

import (
	"sync"

	"github.com/Grivn/phalanx/common/protos"

	"github.com/google/btree"
)

type syncTree struct {
	id    uint64
	mutex sync.Mutex
	qc    *btree.BTree
}

func NewSyncTree(id uint64) *syncTree {
	return &syncTree{id: id, qc: btree.New(2)}
}

func (st *syncTree) SubID() uint64 {
	return st.id
}

func (st *syncTree) Insert(qc *protos.QuorumCert) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.qc.ReplaceOrInsert(qc)
}

func (st *syncTree) Has(qc *protos.QuorumCert) bool {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	return st.qc.Has(qc)
}

func (st *syncTree) Min() *protos.QuorumCert {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	item := st.qc.Min()

	qc, ok := item.(*protos.QuorumCert)
	if !ok {
		return nil
	}

	return qc
}

func (st *syncTree) PullMin() *protos.QuorumCert {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	item := st.qc.DeleteMin()

	qc, ok := item.(*protos.QuorumCert)
	if !ok {
		return nil
	}

	return qc
}

func (st *syncTree) Delete(qc *protos.QuorumCert) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.qc.Delete(qc)
}
