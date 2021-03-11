package logmgr

import (
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
)

func NewLogManager() api.LogManager {
	return newLogMgrImpl()
}

func (lm *logMgrImpl) Save(msg *commonProto.OrderedMsg) {

}
