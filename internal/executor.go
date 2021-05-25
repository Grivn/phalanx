package internal

import commonProto "github.com/Grivn/phalanx/common/protos"

type Executor interface {
	Basic

	ExecuteLogs(exec *commonProto.ExecuteLogs)
}
