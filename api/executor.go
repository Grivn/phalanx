package api

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type Executor interface {
	Basic

	ExecuteLogs(exec *commonProto.ExecuteLogs)
}
