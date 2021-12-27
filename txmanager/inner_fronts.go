package txmanager

import (
	"sync"

	"github.com/Grivn/phalanx/common/protos"
)

type innerFronts struct {
	mutex sync.RWMutex

	fronts map[uint64]*protos.CommandProtoIndex
}

func newInnerFronts() *innerFronts {
	return &innerFronts{fronts: make(map[uint64]*protos.CommandProtoIndex)}
}

func (inner *innerFronts) update(command *protos.Command) {
	inner.mutex.Lock()
	defer inner.mutex.Unlock()

	inner.fronts[command.Author] = &protos.CommandProtoIndex{Author: command.Author, Sequence: command.Sequence}
}

func (inner *innerFronts) read(author uint64) *protos.CommandProtoIndex {
	inner.mutex.RLock()
	defer inner.mutex.RUnlock()
	return inner.fronts[author]
}
