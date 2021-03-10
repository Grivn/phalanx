package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type LogPool interface {
	Save(msg *commonProto.OrderedMsg)

	Load(key interface{}) (*commonProto.OrderedMsg, error)

	Check(key interface{}) bool

	Remove(key interface{})
}
