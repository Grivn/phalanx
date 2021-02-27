package txpool

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

type recorder struct {
	txList   []*commonProto.Transaction
	hashList []string
}

func newRecorder() *recorder {
	return &recorder{
		txList:   nil,
		hashList: nil,
	}
}

func (re *recorder) len() int {
	return len(re.txList)
}

func (re *recorder) txs() []*commonProto.Transaction {
	return re.txList
}

func (re *recorder) hashes() []string {
	return re.hashList
}

func (re *recorder) update(tx *commonProto.Transaction) {
	hash := commonTypes.GetHash(tx)
	re.txList = append(re.txList, tx)
	re.hashList = append(re.hashList, hash)
}

func (re *recorder) reset() {
	re.hashList = nil
	re.txList = nil
}
