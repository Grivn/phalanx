package reliablelog

import (
	"sync"
	"testing"

	"github.com/Grivn/phalanx/api"
	mockapi "github.com/Grivn/phalanx/api/mocks"
	"github.com/Grivn/phalanx/common/mocks"
	types2 "github.com/Grivn/phalanx/common/types"
	"github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/reliablelog/types"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestNewLogManager(t *testing.T) {
	var lms []api.ReliableLog
	var auths []api.Authenticator
	var replyCs []chan types.ReplyEvent
	var netChans []chan interface{}
	ch := make(chan interface{})

	n := 5
	for i:=0; i<n; i++ {
		id := uint64(i+1)
		auth := mockapi.NewAuthMock()
		auths = append(auths, auth)

		logger := mocks.NewRawLogger()

		replyC := make(chan types.ReplyEvent)
		replyCs = append(replyCs, replyC)

		netChan := make(chan interface{})
		netChans = append(netChans, netChan)
		network := mocks.NewReplyNetwork(ch, true)

		lm := NewReliableLog(5, id, replyC, auth, network, logger)
		lms = append(lms, lm)
	}

	for _, lm := range lms{
		lm.Start()
	}

	var bidsset [][]*protos.BatchId
	count := 10
	for index := range lms {
		id := uint64(index+1)
		bids := mocks.NewBatchId(id, count)
		assert.Equal(t, count, len(bids))
		bidsset = append(bidsset, bids)
	}

	// start network dispatcher
	go func() {
		for {
			select {
			case ev := <-ch:
				comm := ev.(*protos.CommMsg)
				for index, netChan := range netChans {
					if index == int(comm.Author-1) {
						continue
					}
					netChan <- comm
				}
			}
		}
	}()

	// wg to control the routine
	var wg sync.WaitGroup

	// wg for quorum events
	wg.Add(n * count)

	// start replicas' network event listener
	for index, lm := range lms {
		go func(lm api.ReliableLog, index int) {
			for {
				select {
				case ev := <-netChans[index]:
					comm := ev.(*protos.CommMsg)
					signed := &protos.SignedMsg{}
					_ = proto.Unmarshal(comm.Payload, signed)
					lm.Record(signed)
				}
			}
		}(lm, index)
	}

	close1 := make(chan bool)
	for index, lm := range lms {
		go func(lm api.ReliableLog, index int) {
			for {
				select {
				case ev := <-replyCs[index]:
					assert.Equal(t, types.LogReplyQuorumBinaryEvent, ev.EventType)
					wg.Done()
				case <-close1:
					return
				}
			}
		}(lm, index)
	}

	for index, lm := range lms {
		go func(lm api.ReliableLog, index int) {
			for _, bid := range bidsset[index] {
				lm.Generate(bid)
			}
		}(lm, index)
	}
	wg.Wait()
	close(close1)

	// wg for execute events
	wg.Add(n * count)

	var tags []*protos.BinaryTag
	for i:=0; i<count; i++ {
		set := []byte{1, 1, 1, 1, 1}
		hash := types2.CalculatePayloadHash(set, 0)
		tag := &protos.BinaryTag{
			Sequence:   uint64(i+1),
			BinaryHash: hash,
			BinarySet:  set,
		}
		tags = append(tags, tag)
	}

	for index, lm := range lms {
		go func(lm api.ReliableLog, index int) {
			for _, tag := range tags {
				lm.Ready(tag)
			}
		}(lm, index)
	}

	for index, lm := range lms {
		go func(lm api.ReliableLog, index int) {
			tagMap := make(map[uint64]bool)
			for {
				select {
				case ev := <-replyCs[index]:
					switch ev.EventType {
					case types.LogReplyQuorumBinaryEvent:
						continue
					case types.LogReplyExecuteEvent:
						exec := ev.Event.(types.ExecuteLogs)
						if !tagMap[exec.Sequence] {
							tagMap[exec.Sequence] = true
							wg.Done()
						}
					case types.LogReplyMissingEvent:
						// todo we need to start a timer for ready event when we use log manager, and stop it when we received the execute event
						continue
					default:
						assert.Fail(t, "invalid type")
						return
					}
				}
			}
		}(lm, index)
	}

	wg.Wait()

	for _, lm := range lms {
		lm.Stop()
	}

	for _, lm := range lms {
		lm.Stop()
	}
}
