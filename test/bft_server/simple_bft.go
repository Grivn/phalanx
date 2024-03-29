package main

import (
	"github.com/Grivn/phalanx/common/protos"
	"sync"
	"time"

	"github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/core"
	"github.com/Grivn/phalanx/external"

	"github.com/google/btree"
)

const (
	proposal = iota
	vote
)

type replica struct {
	mutex sync.Mutex

	author uint64
	quorum int

	phalanx phalanx.Provider

	bftC   chan *bftMessage
	sendC  chan *bftMessage
	closeC chan bool

	sequence uint64
	cache    map[uint64]*bftMessage
	aggMap   map[uint64]int

	executedSeq uint64
	executeCache *btree.BTree

	logger external.Logger
}

type bftMessage struct {
	from     uint64
	to       uint64
	sequence uint64
	digest   string
	typ      int
	pBatch   *protos.PartialOrderBatch
}

func newReplica(n int, author uint64, phx phalanx.Provider, sendC chan *bftMessage, bftC chan *bftMessage, closeC chan bool, logger external.Logger) *replica {
	return &replica{
		quorum:       n-(n-1)/3,
		author:       author,
		phalanx:      phx,
		sendC:        sendC,
		bftC:         bftC,
		closeC:       closeC,
		sequence:     uint64(0),
		cache:        make(map[uint64]*bftMessage),
		aggMap:       make(map[uint64]int),
		executedSeq:  uint64(1),
		executeCache: btree.New(2),
		logger:       logger,
	}
}

func cluster(sendC chan *bftMessage, bftCs map[uint64]chan *bftMessage, closeC chan bool) {
	for {
		select {
		case msg := <-sendC:
			switch msg.to {
			case 0:
				for _, ch := range bftCs {
					ch <- msg
				}
			case 1:
				bftCs[msg.to] <- msg
			default:
				panic("invalid type")
			}
		case <-closeC:
			return
		}
	}
}

func (replica *replica) run() {
	replica.logger.Infof("[%d] running bft consensus", replica.author)

	go replica.bftListener()

	if replica.author == uint64(1) {
		go replica.runningProposal()
	}
}

func (replica *replica) runningProposal() {
	timerC := make(chan bool)
	go func() {
		time.Sleep(500*time.Millisecond)
		timerC <- true
	}()
	for {
		select {
		case <-replica.closeC:
			return
		case <-timerC:
			replica.sequence++
			replica.aggMap[replica.sequence] = 0
			pBatch := &protos.PartialOrderBatch{}
			prop := &bftMessage{from: replica.author, to: 0, typ: proposal, sequence: replica.sequence, digest: types.CalculateBatchHash(pBatch), pBatch: pBatch}
			replica.logger.Infof("[%d] generate proposal sequence %d, hash %s", replica.author, prop.sequence, prop.digest)
			replica.sendC <- prop
			return
		default:
			prop := replica.propose()
			if prop != nil {
				replica.logger.Infof("[%d] generate proposal sequence %d, hash %s", replica.author, prop.sequence, prop.digest)
				replica.sendC <- prop
				return
			}
		}
	}
}

func (replica *replica) bftListener() {
	for {
		select {
		case msg := <-replica.bftC:
			switch msg.typ {
			case proposal:
				replica.processProposal(msg)
			case vote:
				replica.processVote(msg)
			}
		case <-replica.closeC:
			return
		}
	}
}

func (replica *replica) propose() *bftMessage {
	replica.mutex.Lock()
	defer replica.mutex.Unlock()

	if replica.aggMap[replica.sequence] < replica.quorum && replica.sequence != 0 {
		return nil
	}

	pBatch, _ := replica.phalanx.MakeProposal()

	for {
		if pBatch != nil {
			break
		}

		pBatch, _ = replica.phalanx.MakeProposal()
	}

	replica.sequence++
	replica.aggMap[replica.sequence] = 0
	return &bftMessage{from: replica.author, to: 0, typ: proposal, sequence: replica.sequence, digest: types.CalculateBatchHash(pBatch), pBatch: pBatch}
}

func (replica *replica) processProposal(message *bftMessage) *bftMessage {
	replica.mutex.Lock()
	defer replica.mutex.Unlock()

	replica.logger.Infof("[%d] process proposal sequence %d, hash %s", replica.author, message.sequence, message.digest)

	if m, ok := replica.cache[message.sequence-1]; ok && replica.author != uint64(1) {
		replica.execute(m)
	}

	if message.typ != proposal {
		// can only process vote message
		return nil
	}

	if _, ok := replica.cache[message.sequence]; ok {
		// proposed sequence number
		replica.logger.Infof("voted on")
		return nil
	}

	replica.cache[message.sequence] = message
	v := &bftMessage{from: replica.author, to: message.from, typ: vote, sequence: message.sequence}
	go replica.sendVote(v)
	return v
}

func (replica *replica) sendVote(message *bftMessage) {
	replica.logger.Infof("[%d] vote on sequence %d", replica.author, message.sequence)
	replica.sendC <- message
}

func (replica *replica) processVote(message *bftMessage) {
	replica.mutex.Lock()
	defer replica.mutex.Unlock()

	if message.typ != vote {
		// can only process vote messages
		return
	}

	if _, ok := replica.aggMap[message.sequence]; !ok {
		replica.aggMap[message.sequence] = 0
	}

	replica.aggMap[message.sequence]++
	replica.logger.Infof("[%d] process vote for sequence %d, has %d need %d", replica.author, message.sequence, replica.aggMap[message.sequence], replica.quorum)

	if replica.aggMap[message.sequence] == replica.quorum {
		m := replica.cache[message.sequence]
		replica.execute(m)

		go replica.runningProposal()
	}
}

func (replica *replica) execute(message *bftMessage) {
	replica.executeCache.ReplaceOrInsert(message)

	for {
		item := replica.executeCache.Min()

		if item == nil {
			return
		}

		m := item.(*bftMessage)
		if m.sequence != replica.executedSeq {
			return
		}

		replica.logger.Infof("[%d] execute sequence %d, digest %s", replica.author, m.sequence, m.digest)
		if m.pBatch == nil {
			replica.executedSeq++
			continue
		}

		if err := replica.phalanx.CommitProposal(m.pBatch); err != nil {
			panic(err)
		}

		replica.executeCache.Delete(item)
		replica.executedSeq++
	}
}

func (bft *bftMessage) Less(item btree.Item) bool {
	return bft.sequence < (item.(*bftMessage)).sequence
}
