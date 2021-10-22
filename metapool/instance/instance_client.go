package instance

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/external"
	"github.com/Grivn/phalanx/internal"

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
	commandStream *btree.BTree

	//============================ communication channel ========================================

	// receiveC is used to receive command.
	receiveC chan *types.CommandIndex

	// commandC is used to propose command towards log-manager.
	commandC chan *types.CommandIndex

	// committedC is used to receive the committed seqNo.
	committedC chan uint64

	// closeC is used to stop the listener for current client.
	closeC chan bool

	//============================== external interfaces =======================================

	// logger is used to print logs.
	logger external.Logger
}

func NewClient(author, id uint64, commandC chan *types.CommandIndex, logger external.Logger) internal.ClientInstance {
	logger.Infof("[%d] initiate manager for client %d", author, id)
	committedNo := make(map[uint64]bool)
	committedNo[uint64(0)] = true
	return &clientInstance{
		author: author,
		id:     id,

		proposedNo:    uint64(0),
		committedNo:   uint64(0),
		commandStream: btree.New(2),

		receiveC:   make(chan *types.CommandIndex, 1000),
		committedC: make(chan uint64),
		commandC:   commandC,
		closeC:     make(chan bool),

		logger: logger,
	}
}

//======================================= interfaces of client instance =======================================

func (client *clientInstance) Run() {
	client.start()
}

func (client *clientInstance) Close() {
	client.stop()
}

func (client *clientInstance) Commit(seqNo uint64) {
	client.committedC <- seqNo
}

func (client *clientInstance) Append(command *protos.Command) {
	cIndex := types.NewCommandIndex(command)

	client.receiveC <- cIndex
}

//========================================= implement of client instance ===============================================

func (client *clientInstance) start() {
	go client.listener()
}

func (client *clientInstance) stop() {
	select {
	case <-client.closeC:
	default:
		close(client.closeC)
	}
}

func (client *clientInstance) listener() {
	for {
		select {
		case <-client.closeC:
			return
		case cIndex := <-client.receiveC:
			client.commandStream.ReplaceOrInsert(cIndex)
			client.logger.Debugf("[%d] received command %s", client.author, cIndex.Format())
			if c := client.minCommand(); c != nil {
				client.feedBack(c)
			}
		case committed := <-client.committedC:
			client.logger.Debugf("[%d] client %d committed sequence number %d", client.author, client.id, committed)

			if committed != client.committedNo+1 {
				client.logger.Errorf("[%d] invalid committed sequence number, expect %d, committed %d", client.committedNo+1, committed)
			}

			client.committedNo = maxUint64(client.committedNo, committed)

			if c := client.minCommand(); c != nil {
				client.feedBack(c)
			}
		}
	}
}

func (client *clientInstance) minCommand() *types.CommandIndex {

	if client.committedNo < client.proposedNo {
		return nil
	}

	item := client.commandStream.Min()
	if item == nil {
		return nil
	}

	cIndex, ok := item.(*types.CommandIndex)
	if !ok {
		return nil
	}

	if cIndex.SeqNo == client.proposedNo+1 {
		client.commandStream.Delete(item)
		client.proposedNo++
		return cIndex
	}
	return nil
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
