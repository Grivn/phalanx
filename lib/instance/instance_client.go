package instance

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Grivn/phalanx/common/api"
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/google/btree"
)

// clientInstance is used to process transactions from specific client.
//
// we would like to generate an instance for each specific client.
// the client in phalanx naive demo is distinguished with identifier number.
// we should process transactions from each client by order,
// which means we should process (n+1)th tx should be processed
// after the procession for (n)th tx has been finished.
type clientInstance struct {
	// mutex is used to resolve concurrency problems.
	mutex sync.Mutex

	//=============================== basic information =================================

	// author indicates the consensus node identifier.
	author uint64

	// id indicates the identifier for client.
	id uint64

	//=========================== command stream management ==================================

	// proposedNo indicates the proposed seqNo for current client.
	proposedNo uint64

	// committedNo indicates the committed seqNo for current client.
	committedNo uint64

	// commands is used to record the command according to its indicator.
	commands *btree.BTree

	//============================ communication channel ========================================

	// commandC is used to propose command towards log-manager.
	commandC chan<- *types.CommandIndex

	// isActive indicates current client instance's status.
	// if it is true, there is a command from current client waiting to be processed.
	isActive bool

	// activeCount indicates the number of active client instance.
	activeCount *int64

	//
	timestamp int64

	//============================== external interfaces =======================================

	// logger is used to print logs.
	logger external.Logger
}

func NewRelay(author, id uint64, commandC chan<- *types.CommandIndex, activeCount *int64, logger external.Logger) api.Relay {
	logger.Infof("[%d] initiate manager for client %d", author, id)
	committedNo := make(map[uint64]bool)
	committedNo[uint64(0)] = true
	return &clientInstance{
		author:      author,
		id:          id,
		proposedNo:  uint64(0),
		committedNo: uint64(0),
		commands:    btree.New(2),
		commandC:    commandC,
		isActive:    false,
		activeCount: activeCount,
		logger:      logger,
	}
}

func NewClient(author, id uint64, commandC chan<- *types.CommandIndex, activeCount *int64, logger external.Logger) api.ClientInstance {
	logger.Infof("[%d] initiate manager for client %d", author, id)
	committedNo := make(map[uint64]bool)
	committedNo[uint64(0)] = true
	return &clientInstance{
		author:      author,
		id:          id,
		proposedNo:  uint64(0),
		committedNo: uint64(0),
		commands:    btree.New(2),
		commandC:    commandC,
		isActive:    false,
		activeCount: activeCount,
		logger:      logger,
	}
}

func (client *clientInstance) Commit(seqNo uint64) int {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	client.logger.Debugf("[%d] client %d committed sequence number %d", client.author, client.id, seqNo)

	if seqNo != client.committedNo+1 {
		client.logger.Errorf("[%d] invalid committed sequence number, expect %d, committed %d", client.committedNo+1, seqNo)
	}
	client.committedNo = maxUint64(client.committedNo, seqNo)

	if client.commands.Len() == 0 {
		client.hibernate()
	}

	return client.commands.Len()
}

func (client *clientInstance) Append(command *protos.Command) int {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	cIndex := types.NewCommandIndex(command)

	client.commands.ReplaceOrInsert(cIndex)
	client.logger.Debugf("[%d] received command %s", client.author, cIndex.Format())

	c := client.minCommand()

	for {
		if c == nil {
			break
		}

		// the timestamp for partial ordering.
		for {
			current := time.Now().UnixNano()
			if current > client.timestamp {
				c.OTime = time.Now().UnixNano()
				client.timestamp = current
				break
			}
		}

		client.feedBack(c)
		client.activate()
		c = client.minCommand()
	}

	return client.commands.Len()
}

func (client *clientInstance) minCommand() *types.CommandIndex {
	item := client.commands.Min()
	if item == nil {
		return nil
	}

	cIndex, ok := item.(*types.CommandIndex)
	if !ok {
		return nil
	}

	if cIndex.SeqNo == client.proposedNo+1 {
		client.commands.Delete(item)
		client.proposedNo++
		return cIndex
	}
	return nil
}

func (client *clientInstance) activate() {
	if client.isActive {
		return
	}
	val := atomic.AddInt64(client.activeCount, 1)
	client.isActive = true
	client.logger.Debugf("[%d] activate client %d, total active instance %d", client.author, client.id, val)
}

func (client *clientInstance) hibernate() {
	if !client.isActive {
		return
	}
	val := atomic.AddInt64(client.activeCount, -1)
	client.isActive = false
	client.logger.Debugf("[%d] hibernate client %d, total active instance %d", client.author, client.id, val)
}

func (client *clientInstance) feedBack(cIndex *types.CommandIndex) {
	client.commandC <- cIndex
}

func maxUint64(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
