package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type Executor interface {
	Basic

	ExecuteLogs(sequence uint64, logs []*commonProto.OrderedMsg)

	ExecuteBatch(batch *commonProto.Batch)
}
