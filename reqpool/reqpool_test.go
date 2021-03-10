package reqpool

import (
	"sync"
	"testing"

	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/common/utils"

	"github.com/stretchr/testify/assert"
)

func TestRequestPoolImpl_ID(t *testing.T) {
	logger := utils.NewRawLogger()
	replyC := make(chan *commonProto.BatchId)
	rp := NewRequestPool(uint64(1), replyC, logger)
	assert.Equal(t, uint64(1), rp.ID())
}

func TestNewRequestPool(t *testing.T) {
	rps := make(map[uint64]api.RequestPool)
	logger := utils.NewRawLogger()

	replyC := make(chan *commonProto.BatchId)
	closeC := make(chan bool)

	for index:= 0; index <4; index++ {
		id := uint64(index+1)
		rps[id] = NewRequestPool(id, replyC, logger)
	}

	for _, rp := range rps {
		rp.Start()
	}

	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case <-replyC:
				wg.Done()
			case <-closeC:
				return
			}
		}
	}()

	var reqs []*commonProto.OrderedMsg
	for index:= 0; index <4; index++ {
		id := uint64(index + 1)
		tmp := utils.NewOrderedRequests(id, 1, 20)
		reqs = append(reqs, tmp...)
	}
	utils.Shuffle(reqs)

	for _, req := range reqs {
		wg.Add(1)
		rps[req.Author].Record(req)
	}

	wg.Wait()
	close(closeC)

	wg.Add(len(rps))
	for _, rp := range rps {
		rp.Stop()
		wg.Done()
	}
	wg.Wait()

	// stop twice not panic
	for _, rp := range rps {
		rp.Stop()
	}
}
