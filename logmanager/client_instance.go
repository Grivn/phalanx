package logmanager

import (
	"sync"

	"github.com/Grivn/phalanx/common/protos"
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
	mutex sync.Mutex

	h uint64

	// author indicates the consensus node identifier.
	author uint64

	// id indicates the identifier for client.
	id uint64

	// proposedNo indicates the proposed seqNo for current client.
	proposedNo uint64

	// committedNo indicates the committed seqNo for current client.
	committedNo uint64

	receivedCommand map[uint64]*protos.Command

	// commands is used to record the command according to its indicator.
	commands *btree.BTree

	// receiveC is used to receive command.
	receiveC chan *protos.Command

	// commandC is used to propose command towards log-manager.
	commandC chan *protos.Command

	// committedC is used to receive the committed seqNo.
	committedC chan uint64

	// closeC is used to stop the listener for current client.
	closeC chan bool

	// logger is used to print logs.
	logger external.Logger
}

func newClient(author, id uint64, commandC chan *protos.Command, logger external.Logger) *clientInstance {
	logger.Infof("[%d] initiate manager for client %d", author, id)
	committedNo := make(map[uint64]bool)
	committedNo[uint64(0)] = true
	return &clientInstance{
		author: author,
		id: id,
		proposedNo: uint64(0),
		commands: btree.New(2),
		receiveC: make(chan *protos.Command, 1000),
		committedC: make(chan uint64),
		commandC: commandC,
		committedNo: uint64(0),
		logger: logger,
		receivedCommand: make(map[uint64]*protos.Command),
	}
}

func (client *clientInstance) start() {
	go client.listener()
}

func (client *clientInstance) listener() {
	for {
		select {
		case <-client.closeC:
			return
		case command := <-client.receiveC:
			client.commands.ReplaceOrInsert(command)
			client.logger.Debugf("[%d] received command %s", client.author, command.Format())
			if c := client.minCommand(); c != nil {
				go client.feedBack(c)
			}
		case committed := <-client.committedC:
			client.logger.Debugf("[%d] client %d committed sequence number %d", client.author, client.id, committed)

			if committed != client.committedNo+1 {
				panic("invalid committed number")
			}

			client.committedNo = committed

			if c := client.minCommand(); c != nil {
				go client.feedBack(c)
			}
		}
	}
}

func (client *clientInstance) commit(seqNo uint64) {
	client.committedC <- seqNo
	//client.mutex.Lock()
	//defer client.mutex.Unlock()
	//
	//client.logger.Debugf("[%d] client %d committed sequence number %d", client.author, client.id, seqNo)
	//client.committedNo[seqNo] = true
	//if c := client.minCommand(); c != nil {
	//	go client.feedBack(c)
	//}
}

func (client *clientInstance) append(command *protos.Command) {
	client.receiveC <- command
	//client.mutex.Lock()
	//defer client.mutex.Unlock()
	//
	////client.logger.Debugf("[%d] received command %s", client.author, command.Format())
	//client.commands.ReplaceOrInsert(command)
	//if c := client.minCommand(); c != nil {
	//	go client.feedBack(c)
	//}
}

func (client *clientInstance) minCommand() *protos.Command {

	if client.committedNo < client.proposedNo {
		return nil
	}

	item := client.commands.Min()
	if item == nil {
		return nil
	}

	command, ok := item.(*protos.Command)
	if !ok {
		return nil
	}

	if command.Sequence == client.proposedNo+1 {
		client.commands.Delete(item)
		client.proposedNo++
		return command
	}
	return nil
}

func (client *clientInstance) feedBack(command *protos.Command) {
	client.commandC <- command
}
