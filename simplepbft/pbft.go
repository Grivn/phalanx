package simplepbft

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type pbft struct {
	author uint64

	certs map[uint64]*cert

	check map[uint64]*checkpoint

	tags map[uint64]*commonProto.BinaryTag

	h uint64

	seqNo uint64

	close chan bool

	logger external.Logger
}

type cert struct {
	prePreps interface{}

	preps map[uint64]interface{}

	commit map[uint64]interface{}
}

type checkpoint struct {
	checkpoint map[uint64]interface{}
}

func (p *pbft) start() {
	go p.listener()
}

func (p *pbft) stop() {
	select {
	case <-p.close:
	default:
		close(p.close)
	}
}

func (p *pbft) listener() {

}

func (p *pbft) processEvent() {

}

func (p *pbft) recvTags() {

}

func (p *pbft) sendPrePrepare() {

}

func (p *pbft) recvPrePrepare() {

}

func (p *pbft) recvPrepare() {

}

func (p *pbft) recvCommit() {

}

func (p *pbft) recvCheckpoint() {

}
