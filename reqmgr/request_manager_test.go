package reqmgr

import (
	"github.com/Grivn/phalanx/reqmgr/types"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"

	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/common/utils"
)

func TestNewRequestPool(t *testing.T) {
	logger := utils.NewRawLogger()
	network := utils.NewFakeNetwork()

	replyC := make(chan interface{})
	closeC := make(chan bool)

	n := 5
	author := uint64(1)
	rm := NewRequestManager(n, author, replyC, network, logger)
	rm.Start()

	var reqs []*commonProto.OrderedMsg
	for index:= 0; index <n; index++ {
		id := uint64(index + 1)
		tmp := utils.NewOrderedMessages(id, 1, 20, commonProto.OrderType_REQ)
		reqs = append(reqs, tmp...)
	}
	utils.Shuffle(reqs)

	var wg sync.WaitGroup
	wg.Add(len(reqs))

	go func() {
		for {
			select {
			case ev := <-replyC:
				event := ev.(types.ReplyEvent)
				assert.Equal(t, types.ReqReplyBatchByOrder, event.EventType)
				wg.Done()
			case <-closeC:
				return
			}
		}
	}()

	go func() {
		for _, req := range reqs {
			rm.Record(req)
		}
	}()

	wg.Wait()
	close(closeC)

	rm.Stop()

	// stop twice not panic
	rm.Stop()
}
