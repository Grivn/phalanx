package requester

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type recvChan struct {
	orderedChan chan *commonProto.OrderedMsg
}

type sendChan struct {
	batchIdChan chan *commonProto.BatchId
}
