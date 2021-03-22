package phalanx

import (
	"strconv"
	"sync"
	"testing"

	"github.com/Grivn/phalanx"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/types/protos"
)

func TestPhalanx(t *testing.T) {
	var phs []phalanx.Phalanx
	var netChans []chan interface{}
	ch := make(chan interface{})

	n := 5
	for i:=0; i<n; i++ {
		id := uint64(i+1)

		logger := mocks.NewRawLoggerFile("node"+strconv.Itoa(int(id)))

		netChan := make(chan interface{})
		netChans = append(netChans, netChan)
		network := mocks.NewReplyNetwork(ch, true)

		exec := mocks.NewSimpleExecutor(id, logger)

		ph := NewPhalanx(n, id, exec, network, logger)
		phs = append(phs, ph)
	}

	for _, ph := range phs {
		ph.Start()
	}

	count := 1000

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

	// start replicas' network event listener
	for index, ph := range phs {
		go func(ph phalanx.Phalanx, index int) {
			for {
				select {
				case ev := <-netChans[index]:
					comm := ev.(*protos.CommMsg)
					ph.Propose(comm)
				}
			}
		}(ph, index)
	}

	var wg sync.WaitGroup

	wg.Add(n+1)
	for _, ph := range phs {
		go func(ph phalanx.Phalanx) {
			var txs []*protos.Transaction
			for i:=0; i<count; i++ {
				tx := mocks.NewTx()
				txs = append(txs, tx)
			}
			ph.PostTxs(txs)
			wg.Done()
		}(ph)
	}
	wg.Wait()
}
