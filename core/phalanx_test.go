package phalanx

import (
	"strconv"
	"testing"
	"time"

	"github.com/Grivn/phalanx"
	"github.com/Grivn/phalanx/api"
	mockapi "github.com/Grivn/phalanx/api/mocks"
	"github.com/Grivn/phalanx/common/mocks"
	"github.com/Grivn/phalanx/common/types/protos"

	"github.com/golang/mock/gomock"
)

func TestPhalanx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	var phs []phalanx.Phalanx
	var auths []api.Authenticator
	var netChans []chan interface{}
	ch := make(chan interface{})

	n := 5
	for i:=0; i<n; i++ {
		id := uint64(i+1)

		auth := mockapi.NewAuthenticatorMinimal(ctrl)
		auths = append(auths, auth)

		logger := mocks.NewRawLoggerFile("node"+strconv.Itoa(int(id)))

		netChan := make(chan interface{})
		netChans = append(netChans, netChan)
		network := mocks.NewReplyNetwork(ch, false)

		exec := mocks.NewSimpleExecutor(id, logger)

		ph := NewPhalanx(n, id, auth, exec, network, logger)
		phs = append(phs, ph)
	}

	for _, ph := range phs {
		ph.Start()
	}

	count := 20000

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

	for _, ph := range phs {
		go func(ph phalanx.Phalanx) {
			var txs []*protos.Transaction
			for i:=0; i<count; i++ {
				if ph.IsNormal() {
					tx := mocks.NewTx()
					txs = append(txs, tx)
				} else {
					i--
				}
			}
			ph.PostTxs(txs)
		}(ph)
	}

	time.Sleep(5*time.Second)
	for _, ph := range phs {
		ph.Stop()
	}
	time.Sleep(1*time.Second)
}
