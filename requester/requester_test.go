package requester
//
//import (
//	"sync"
//	"testing"
//
//	"github.com/Grivn/phalanx/common/mocks"
//	commonProto "github.com/Grivn/phalanx/common/types/protos"
//)
//
//func TestNewRequestPool(t *testing.T) {
//	logger := mocks.NewRawLogger()
//	network := mocks.NewFakeNetwork()
//
//	bidC := make(chan *commonProto.BatchId)
//	closeC := make(chan bool)
//
//	n := 5
//	author := uint64(1)
//	rm := NewRequester(n, author, bidC, network, logger)
//	rm.Start()
//
//	var reqs []*commonProto.OrderedMsg
//	for index:= 0; index <n; index++ {
//		id := uint64(index + 1)
//		tmp := mocks.NewOrderedMessages(id, 1, 20, commonProto.OrderType_REQ)
//		reqs = append(reqs, tmp...)
//	}
//	mocks.Shuffle(reqs)
//
//	var wg sync.WaitGroup
//	wg.Add(len(reqs))
//
//	go func() {
//		for {
//			select {
//			case <-bidC:
//				wg.Done()
//			case <-closeC:
//				return
//			}
//		}
//	}()
//
//	go func() {
//		for _, req := range reqs {
//			rm.Record(req)
//		}
//	}()
//
//	wg.Wait()
//	close(closeC)
//
//	rm.Stop()
//
//	// stop twice not panic
//	rm.Stop()
//}
