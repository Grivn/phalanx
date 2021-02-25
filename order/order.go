package order

import (
	commonBasic "github.com/Grivn/phalanx/common/basic"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

// Order
type Order interface {
	 commonBasic.Basic

	 // only receive the tx hash from api
	 // we would like to generate an ordered-req if we have received efficient txs or timeout
	 // api tx --> order ---> sgx ---> ordered-req
	 ReceiveTransaction(txs *commonProto.TransactionSet)

	 ReceiveOrderedMsg(msg *commonProto.OrderedMsg)
}
