package logmgr

import (
	"errors"
	"github.com/Grivn/phalanx/api"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/common/utils"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewLogPool(t *testing.T) {
	lps := make(map[uint64]api.LogPool)
	logger := utils.NewRawLogger()

	for index:= 0; index <4; index++ {
		id := uint64(index+1)
		lps[id] = NewLogPool(id, logger)
	}

	var logs []*commonProto.OrderedMsg
	for index:= 0; index <4; index++ {
		id := uint64(index + 1)
		tmp := utils.NewOrderedMessages(id, 1, 50, commonProto.OrderType_REQ)
		logs = append(logs, tmp...)
	}

	assert.Equal(t, uint64(1), lps[uint64(1)].ID())

	ret, err := lps[uint64(1)].Load(uint64(1))
	assert.Nil(t, ret)
	assert.Equal(t, errors.New("cannot find log"), err)

	var wg sync.WaitGroup
	wg.Add(len(logs))
	for _, log := range logs {
		go func(log *commonProto.OrderedMsg) {
			lps[log.Author].Save(log)
			v1, err := lps[log.Author].Load(log.Sequence)
			assert.Nil(t, err)
			assert.Equal(t, log, v1)
			v2, err := lps[log.Author].Load(log.BatchId.BatchHash)
			assert.Nil(t, err)
			assert.Equal(t, log, v2)
			assert.True(t, lps[log.Author].Check(log.Sequence))
			assert.True(t, lps[log.Author].Check(log.BatchId.BatchHash))
			lps[log.Author].Remove(log)
			wg.Done()
		}(log)
	}


	wg.Wait()
}
