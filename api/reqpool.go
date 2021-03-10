package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

// we would like to generate a request pool for every replica in cluster
// to process the ordered requests send from each of them
type RequestPool interface {
	Basic

	ID() uint64

	Record(msg *commonProto.OrderedMsg)
}
