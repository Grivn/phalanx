package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type BinaryByzantine interface {
	Basic

	Trigger(tag *commonProto.BinaryTag)
}
