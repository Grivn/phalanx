package cmdmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"sync"
)

// cmdManager is the implement of phalanx client, which is used for performance basic test.
type cmdManager struct {
	mutex sync.Mutex

	n int

	author uint64

	interval uint64

	seqNo uint64

	commandSize int

	txSet []*protos.Transaction

	sender external.TestSender

	logger external.Logger
}

func NewTestReceiver(n int, author uint64, commandSize int, sender external.TestSender, logger external.Logger) *cmdManager {
	return &cmdManager{n: n, author: author, interval: author, commandSize: commandSize, sender: sender, logger: logger}
}

func (cmd *cmdManager) ProcessTransaction(tx *protos.Transaction) {
	cmd.mutex.Lock()
	defer cmd.mutex.Unlock()

	cmd.txSet = append(cmd.txSet, tx)
	if len(cmd.txSet) == cmd.commandSize {
		cmd.seqNo++
		command := types.GenerateCommand(cmd.author, cmd.seqNo, cmd.txSet)
		//command := types.GenerateCommand(uint64(1), (cmd.seqNo-1)*uint64(cmd.n)+cmd.interval, cmd.txSet)
		cmd.sender.BroadcastCommand(command)
		cmd.logger.Infof("[%d] generate command %s", cmd.author, command.Format())
		cmd.txSet = nil
	}
}
