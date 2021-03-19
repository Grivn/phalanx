package binbyzantine

import (
	"github.com/Grivn/phalanx/api"
	"github.com/Grivn/phalanx/binbyzantine/types"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/types/protos"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewByzantine(t *testing.T) {
	n := 5

	count := 6

	var bbas []api.BinaryByzantine


	var replyCs []chan types.ReplyEvent

	var netChans []chan interface{}
	ch := make(chan interface{})

	for i:=0; i<n; i++ {
		id := uint64(i+1)

		replyC := make(chan types.ReplyEvent)
		replyCs = append(replyCs, replyC)

		netChan := make(chan interface{})
		netChans = append(netChans, netChan)
		network := mocks.NewReplyNetwork(ch, true)

		logger := mocks.NewRawLogger()

		bba := NewByzantine(n, id, replyC, network, logger)
		bbas = append(bbas, bba)
	}

	for _, bba := range bbas {
		bba.Start()
	}

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

	// start replicas' network event listener
	for index, bba := range bbas {
		go func(bba api.BinaryByzantine, index int) {
			for {
				select {
				case ev := <-netChans[index]:
					comm := ev.(*protos.CommMsg)
					ntf := &protos.BinaryNotification{}
					_ = proto.Unmarshal(comm.Payload, ntf)
					time.Sleep(1*time.Millisecond)
					bba.Propose(ntf)
				}
			}
		}(bba, index)
	}

	var wg sync.WaitGroup

	var tagss [][]*protos.BinaryTag

	wg.Add(n * count)

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		tags := mocks.NewBinaryTag(n, id, 1, uint64(count))
		tagss = append(tagss, tags)
	}

	for index, bba := range bbas {
		go func(bba api.BinaryByzantine, index int) {
			for _, tag := range tagss[index] {
				bba.Trigger(tag)
			}
		}(bba, index)
	}

	close1 := make(chan bool)
	for index, bba := range bbas {
		go func(bba api.BinaryByzantine, index int) {
			for {
				select {
				case ev := <-replyCs[index]:
					assert.Equal(t, types.BinaryReplyReady, ev.EventType)
					wg.Done()
				case <-close1:
					return
				}
			}
		}(bba, index)
	}
	wg.Wait()

	for _, bba := range bbas {
		bba.Stop()
	}

	for _, bba := range bbas {
		bba.Stop()
	}
}
