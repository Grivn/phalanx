package logmanager

import (
	"github.com/Grivn/phalanx/common/protos"
	"github.com/Grivn/phalanx/external"
	"sync"
)

// commandTracker is used to record the commands current node has received.
// 1) receive commands from clients directly.
// 2) receive commands with the fetch-missing process.
// the tracker of commands belongs to log manager.
type commandTracker struct {
	// mutex is used to control the concurrency problems of command tracker.
	mutex sync.Mutex

	// author indicates current node identifier.
	author uint64

	// commandMap records the commands current node has received.
	commandMap map[string]*protos.Command

	// committedMap records the commands which have been committed.
	committedMap map[string]bool

	// logger prints logs.
	logger external.Logger
}

func newCommandTracker(author uint64, logger external.Logger) *commandTracker {
	logger.Infof("[%d] initiate command tracker", author)
	return &commandTracker{
		author:       author,
		commandMap:   make(map[string]*protos.Command),
		committedMap: make(map[string]bool),
		logger:       logger,
	}
}

func (ct *commandTracker) recordCommand(command *protos.Command) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	if _, ok := ct.commandMap[command.Digest]; ok {
		// duplicated command.
		ct.logger.Debugf("[%d] duplicated command %s", ct.author, command.Digest)
		return
	}

	if ct.committedMap[command.Digest] {
		// committed command
		ct.logger.Debugf("[%d] committed command %s", ct.author, command.Digest)
		return
	}

	ct.logger.Debugf("[%d] received command %s", ct.author, command.Digest)
	ct.commandMap[command.Digest] = command
}

func (ct *commandTracker) readCommand(digest string) *protos.Command {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	command, ok := ct.commandMap[digest]
	if !ok {
		return nil
	}
	delete(ct.commandMap, digest)
	return command
}

func (ct *commandTracker) commitCommand(digest string) {
	ct.mutex.Lock()
	defer ct.mutex.Unlock()

	ct.logger.Debugf("[%d] commit command %s", ct.author, digest)
	delete(ct.commandMap, digest)
	ct.committedMap[digest] = true
}