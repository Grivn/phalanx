package logmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"sync"
)

type chainedManager struct {
	//===================================== basic information =========================================

	// mutex is used for concurrency control of current chained manager.
	mutex sync.Mutex

	// author indicates current node identifier.
	author uint64

	// highOrder is the highest partial order for current chained manager.
	highOrder *protos.PartialOrder

	//===================================== client commands manager ============================================

	// clients track the commands from them.
	clients map[uint64]*clientInstance

	// commandC is used to receive the valid transaction from one client instance.
	commandC chan *protos.Command

	// closeC is used to stop log manager.
	closeC chan bool

	//======================================= external tools ===========================================

	// sender is used to send consensus message into network.
	sender external.NetworkService

	// logger is used to print logs.
	logger external.Logger
}
