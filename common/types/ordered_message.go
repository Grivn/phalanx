package types

import (
	"time"

	"github.com/Grivn/phalanx/common/types/protos"
)

func GenerateTxListHash(txList []*protos.Transaction) []byte {
	var hashList []string
	for _, tx := range txList {
		hashList = append(hashList, GetHash(tx))
	}
	return CalculateMD5Hash(hashList, time.Now().UnixNano())
}

func GenerateOrderedLog(author uint64, sequence uint64) {

}
