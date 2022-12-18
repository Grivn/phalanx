package memorypool

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"sync"

	"github.com/Grivn/phalanx/metrics"
	"github.com/Grivn/phalanx/pkg/common/api"
	"github.com/Grivn/phalanx/pkg/common/config"
	"github.com/Grivn/phalanx/pkg/common/protos"
	"github.com/Grivn/phalanx/pkg/common/types"
	"github.com/Grivn/phalanx/pkg/external"
	"github.com/Grivn/phalanx/pkg/utils/instance"
)

type memoryPoolImpl struct {
	//===================================== basic information =========================================

	// mutex is used to deal with the concurrent problems of log-manager.
	mutex sync.RWMutex

	// author is the identifier for current node.
	author uint64

	// n indicates the number of participants in current cluster.
	n int

	// multi indicates the number of proposers each node maintains.
	multi int

	//==================================== sub-chain management =============================================

	// quorum is the legal size for current node.
	quorum int

	// commandSet is used to record the commands' waiting list according to receive order.
	commandSet types.CommandSet

	//===================================== client commands manager ============================================

	//=================================== local timer service ========================================

	// timer is used to control the timeout event to generate order with commands in waiting list.
	timer api.SingleTimer

	sequencerInstanceMap map[uint64]api.SequencerInstance

	commandTracker api.CommandTracker

	attemptTracker api.AttemptTracker

	checkpointTracker api.CheckpointTracker

	consensusEngine api.ConsensusEngine

	crypto api.CryptoService

	//======================================= consensus manager ============================================

	// commitNo indicates the maximum committed number for each participant's partial order.
	commitNo map[uint64]uint64

	// metrics is used to record the metric info of current node's meta pool.
	metrics *metrics.MetaPoolMetrics

	logger external.Logger
}

func NewMemoryPool(
	conf config.PhalanxConf,
	engine api.ConsensusEngine,
	commandTracker api.CommandTracker,
	attemptTracker api.AttemptTracker,
	checkpointTracker api.CheckpointTracker,
	cryptoService api.CryptoService,
	logger external.Logger) *memoryPoolImpl {
	logger.Infof("[%d] initiate log manager, replica count %d", conf.NodeID, conf.NodeCount)
	return &memoryPoolImpl{
		author:               conf.NodeID,
		n:                    conf.NodeCount,
		multi:                conf.Multi,
		quorum:               types.CalculateQuorum(conf.NodeCount),
		consensusEngine:      engine,
		commandTracker:       commandTracker,
		attemptTracker:       attemptTracker,
		checkpointTracker:    checkpointTracker,
		crypto:               cryptoService,
		sequencerInstanceMap: make(map[uint64]api.SequencerInstance),
		commitNo:             make(map[uint64]uint64),
		logger:               logger,
	}
}

func (mp *memoryPoolImpl) ProcessCommand(command *protos.Command) {
	mp.commandTracker.Record(command)
}

func (mp *memoryPoolImpl) ProcessOrderAttempt(attempt *protos.OrderAttempt) {
	mp.attemptTracker.Record(attempt)

	sequencerInstance, ok := mp.sequencerInstanceMap[attempt.NodeID]
	if !ok {
		sequencerInstance = instance.NewSequencerInstance(mp.author, attempt.NodeID, mp.logger)

		mp.mutex.Lock()
		mp.sequencerInstanceMap[attempt.NodeID] = sequencerInstance
		mp.mutex.Unlock()
	}

	if attempt == nil {
		return
	}
	sequencerInstance.Append(attempt)
}

func (mp *memoryPoolImpl) ProcessConsensusMessage(consensusMessage *protos.ConsensusMessage) {
	switch consensusMessage.Type {
	case protos.MessageType_ORDER_ATTEMPT:
		attemtp := &protos.OrderAttempt{}
		_ = proto.Unmarshal(consensusMessage.Payload, attemtp)
		go mp.ProcessOrderAttempt(attemtp)
		return
	default:
		mp.consensusEngine.ProcessConsensusMessage(consensusMessage)
	}
}

func (mp *memoryPoolImpl) GenerateProposal() (*protos.Proposal, error) {
	mp.mutex.RLock()
	defer mp.mutex.RUnlock()
	proposal := protos.NewProposal(mp.author, mp.n)

	for id, sequencerInstance := range mp.sequencerInstanceMap {
		index := int(id - 1)
		hAttempt := sequencerInstance.GetHighestAttempt()

		if hAttempt == nil {
			// high-order for replica 'id' is nil, record 0 in batch tracker.
			proposal.HighestCheckpointList[index] = protos.NewNopCheckpoint()
			proposal.SeqList[index] = 0
			continue
		}

		idx := types.QueryIndex{Author: hAttempt.NodeID, SeqNo: hAttempt.SeqNo}
		if !mp.checkpointTracker.IsExist(idx) {
			go mp.consensusEngine.ProcessLocalEvent(types.LocalEvent{Type: types.LocalEventOrderAttempt, Event: hAttempt})
		}

		for {
			checkpoint := mp.checkpointTracker.Get(idx)
			if checkpoint != nil {
				// update batch tracker with information of high-order.
				proposal.HighestCheckpointList[index] = checkpoint
				proposal.SeqList[index] = checkpoint.SeqNo()
				break
			}
		}
	}

	mp.logger.Debugf("[%d] generate proposal %s", mp.author, proposal.Format())
	return proposal, nil
}

func (mp *memoryPoolImpl) VerifyProposal(proposal *protos.Proposal) (types.QueryStream, error) {
	updated := false

	for index, no := range proposal.SeqList {

		// calculate the node id.
		id := uint64(index + 1)

		if no <= mp.commitNo[id] {
			// committed previous partial order for node id, including partial number 0.
			mp.logger.Debugf("[%d] haven't updated committed partial order for node %d, no %d", mp.author, id, no)
			continue
		}

		checkpoint := proposal.HighestCheckpointList[index]

		if checkpoint.SeqNo() != no {
			return nil, fmt.Errorf("invalid partial order seqNo, proposedNo %d, partial seqNo %d", no, checkpoint.SeqNo())
		}

		qIndex := types.QueryIndex{Author: checkpoint.NodeID(), SeqNo: checkpoint.SeqNo()}
		if !mp.checkpointTracker.IsExist(qIndex) {
			if err := mp.crypto.VerifyProofCerts(types.StringToBytes(checkpoint.Digest()), checkpoint.QC, mp.quorum); err != nil {
				return nil, fmt.Errorf("invalid high partial order received from %d: %s", proposal.Author, err)
			}
		}

		updated = true
	}

	if !updated {
		// haven't updated committed partial order.
		return nil, nil
	}

	var qStream types.QueryStream

	for index, no := range proposal.SeqList {
		id := uint64(index + 1)

		for {
			if no <= mp.commitNo[id] {
				break
			}

			mp.commitNo[id]++

			qIndex := types.NewQueryIndex(id, mp.commitNo[id])
			qStream = append(qStream, qIndex)
		}
	}

	return qStream, nil
}

//===============================================================
//                   Read Essential Info
//===============================================================

func (mp *memoryPoolImpl) ReadCommand(commandD string) *protos.Command {
	command := mp.commandTracker.Get(commandD)

	for {
		if command != nil {
			break
		}

		// if we could not read the command, just try the next time.
		command = mp.commandTracker.Get(commandD)
	}

	return command
}

func (mp *memoryPoolImpl) ReadOrderAttempts(qStream types.QueryStream) []*protos.OrderAttempt {
	var res []*protos.OrderAttempt

	for _, qIndex := range qStream {
		attempt := mp.attemptTracker.Get(qIndex)

		for {
			if attempt != nil {
				break
			}

			// if we could not read the partial order, just try the next time.
			attempt = mp.attemptTracker.Get(qIndex)
		}

		res = append(res, attempt)
	}

	return res
}
