package order

import commonProto "github.com/Grivn/phalanx-common/types/protos"

// Order
type Order interface {
	 Start()
	 Stop()

	 // only receive the tx hash from api
	 // we would like to generate an ordered-req if we have received efficient txs or timeout
	 // api tx --> order ---> sgx ---> ordered-req
	 ReceiveTransaction(txs *commonProto.TransactionSet)

	 // receive the ordered-req
	 // every ordered-req would be assigned a log sequence number
	 // we would like to generate an ordered-log-set if we have generated efficient ordered-log or timeout
	 // we cannot receive ordered request from 'author == self'
	 // network ---> order ---> sgx(verify) ---> assign seq ---> sgx(sign) ---> seqpool
	 // local ----->
	 ReceiveOrderedReq(req *commonProto.OrderedReq)

	 // receive the ordered-log-set
	// we cannot receive ordered log from 'author == self'
	 // network ---> order ---> sgx(verify) ---> seqpool
	 // local ----->
	 ReceiveOrderedLog(log *commonProto.OrderedLog)
}
