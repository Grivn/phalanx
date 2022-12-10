package instance

import (
	"sync"
	"time"

	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/google/btree"
)

type relayInstance struct {
	// mutex is used to resolve concurrency problems.
	mutex sync.Mutex

	//=============================== basic information =================================

	// sequencerID indicates the consensus node identifier.
	sequencerID uint64

	// relayID indicates the identifier for client.
	relayID uint64

	//=========================== command stream management ==================================

	// proposedNo indicates the proposed seqNo for current client.
	proposedNo uint64

	// commandTree is used to record the command according to its indicator.
	commandTree *btree.BTree

	//============================ communication channel ========================================

	// cIndexC is used to propose command towards log-manager.
	cIndexC chan<- *types.CommandIndex

	//
	timestamp int64

	//============================== external interfaces =======================================

	// logger is used to print logs.
	logger external.Logger
}

func NewRelay(sequencerID, relayID uint64, cIndexC chan<- *types.CommandIndex, logger external.Logger) api.Relay {
	logger.Infof("[%d] initiate manager for relay %d", sequencerID, relayID)
	committedNo := make(map[uint64]bool)
	committedNo[uint64(0)] = true
	return &relayInstance{
		sequencerID: sequencerID,
		relayID:     relayID,
		proposedNo:  uint64(0),
		commandTree: btree.New(2),
		cIndexC:     cIndexC,
		logger:      logger,
	}
}

func (client *relayInstance) Append(command *protos.Command) int {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	cIndex := types.NewCommandIndex(command)

	client.commandTree.ReplaceOrInsert(cIndex)
	client.logger.Debugf("[%d] received command %s", client.sequencerID, cIndex.Format())

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
		c = client.minCommand()
	}

	return client.commandTree.Len()
}

func (client *relayInstance) minCommand() *types.CommandIndex {
	item := client.commandTree.Min()
	if item == nil {
		return nil
	}

	cIndex, ok := item.(*types.CommandIndex)
	if !ok {
		return nil
	}

	if cIndex.SeqNo == client.proposedNo+1 {
		client.commandTree.Delete(item)
		client.proposedNo++
		return cIndex
	}
	return nil
}

func (client *relayInstance) feedBack(cIndex *types.CommandIndex) {
	client.cIndexC <- cIndex
}
