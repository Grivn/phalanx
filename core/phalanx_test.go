package phalanx

import (
	commonTypes "github.com/Grivn/phalanx/common/types"
	phalanxTypes "github.com/Grivn/phalanx/core/types"
	"github.com/Grivn/phalanx/external"
	"github.com/gogo/protobuf/proto"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Grivn/phalanx"
	"github.com/Grivn/phalanx/common/mocks"
	commonProto "github.com/Grivn/phalanx/common/protos"
)

type data struct {
	mutex sync.Mutex

	primary map[uint64]*commonProto.ExecuteLogs

	replicas map[uint64]map[uint64]map[uint64]*commonProto.OrderedLog

	seqNo map[uint64]uint64
}

type testNode struct {
	id uint64

	seqNo uint64

	phalanx phalanx.Phalanx

	commC commonTypes.CommChan

	reliableC commonTypes.ReliableSendChan

	stable map[uint64][]*commonProto.OrderedLog

	data *data

	logger external.Logger
}

func NewTestNode(id uint64, conf phalanxTypes.Config, data *data) *testNode {
	return &testNode{
		id:        id,
		seqNo:     uint64(1),
		phalanx:   NewPhalanx(conf),
		commC:     conf.CommC,
		reliableC: conf.ReliableC,
		data:      data,
		stable:    make(map[uint64][]*commonProto.OrderedLog),
		logger:    mocks.NewRawLogger(),
	}
}

func (test *testNode) listenReliable() {
	for {
		select {
		case log := <-test.reliableC.TrustedChan:
			test.data.mutex.Lock()
			replica := test.data.replicas[test.id]
			sequenced, ok := replica[test.id]
			if !ok {
				sequenced = make(map[uint64]*commonProto.OrderedLog)
				replica[test.id] = sequenced
			}
			sequenced[log.Author] = log
			test.data.mutex.Unlock()
			test.execute()

		case log := <-test.reliableC.StableChan:
			test.stable[log.Sequence] = append(test.stable[log.Sequence], log)
			if len(test.stable[log.Sequence]) >= 3 && test.id == uint64(1) {
				test.data.mutex.Lock()
				test.data.primary[log.Sequence] = &commonProto.ExecuteLogs{
					Sequence: log.Sequence,
					OrderedLogs: test.stable[log.Sequence],
				}
				test.data.mutex.Unlock()
				test.stable[log.Sequence] = nil
			}
			test.execute()
		}
	}
}

func (test *testNode) execute() {
	test.data.mutex.Lock()
	defer test.data.mutex.Unlock()

	for {
		seq := test.data.seqNo[test.id]
		exec, ok := test.data.primary[seq]
		if !ok {
			break
		}
		payload, _ := proto.Marshal(exec)
		test.phalanx.Execute(payload)
		test.data.seqNo[test.id]++
	}
}

func (test *testNode) check(exec *commonProto.ExecuteLogs, trusted map[uint64]*commonProto.OrderedLog) bool {
	for _, log := range exec.OrderedLogs {
		if _, ok := trusted[log.Author]; !ok {
			return false
		}
	}
	return true
}

func TestPhalanx(t *testing.T) {
	n := 4
	ccs := make(map[uint64]commonTypes.CommChan)
	testNodes := make(map[uint64]*testNode)

	data := &data{
		primary:  make(map[uint64]*commonProto.ExecuteLogs),
		replicas: make(map[uint64]map[uint64]map[uint64]*commonProto.OrderedLog),
		seqNo:    make(map[uint64]uint64),
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		ccs[id] = commonTypes.CommChan{
			BatchChan: make(chan *commonProto.Batch),
			ReqChan:   make(chan *commonProto.OrderedReq),
			LogChan:   make(chan *commonProto.OrderedLog),
			AckChan:   make(chan *commonProto.OrderedAck),
		}
		data.replicas[id] = make(map[uint64]map[uint64]*commonProto.OrderedLog)
		data.seqNo[id] = uint64(1)
	}

	for i:=0; i<n; i++ {
		id := uint64(i+1)
		logger := mocks.NewRawLoggerFile("node"+strconv.Itoa(int(id)))
		network := mocks.NewFakeNetwork(id, ccs)
		exec := mocks.NewSimpleExecutor(id, logger)
		rc := commonTypes.ReliableSendChan{
			TrustedChan: make(chan *commonProto.OrderedLog),
			StableChan:  make(chan *commonProto.OrderedLog),
		}
		conf := phalanxTypes.Config{
			N:         n,
			Author:    id,
			BatchSize: 10,
			PoolSize:  50000,
			CommC:     ccs[id],
			ReliableC: rc,
			Executor:  exec,
			Network:   network,
			Logger:    logger,
		}
		testNodes[id] = NewTestNode(id, conf, data)
		go testNodes[id].listenReliable()
	}

	for _, node := range testNodes {
		node.phalanx.Start()
	}

	count := 10000
	for _, node := range testNodes {
		go func(ph phalanx.Phalanx) {
			var txs []*commonProto.Transaction
			for i:=0; i<count; i++ {
				time.Sleep(10*time.Microsecond)
				if ph.IsNormal() {
					tx := mocks.NewTx()
					txs = append(txs, tx)
				} else {
					println("is pool full")
					i--
				}
			}
			ph.PostTxs(txs)
		}(node.phalanx)
	}

	time.Sleep(1000*time.Second)

	for _, node := range testNodes {
		node.phalanx.Stop()
	}
}
