package reliable

import (
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
)

type reliableImpl struct {
	monitored uint64

	sequence uint64

	recvC chan interface{}

	closeC chan bool

	logger external.Logger
}

func newReliableImpl() *reliableImpl {
	return &reliableImpl{
		sequence: uint64(0),
	}
}

func (r *reliableImpl) start() {

}

func (r *reliableImpl) stop() {

}

func (r *reliableImpl) propose(event interface{}) {
	r.recvC <- event
}

func (r *reliableImpl) listener() {
	for {
		select {
		case <-r.closeC:
			r.logger.Notice("exist listener of reliable broadcast")
			return
		case ev := <-r.recvC:
			r.processEvents(ev)
		}
	}
}

func (r *reliableImpl) processEvents(event interface{}) {
	switch event.(type) {
	case *commonProto.ConsensusMessage:
	case *commonProto.Batch:
		r.sendRequest()
	}
}

func (r *reliableImpl) dispatchMessages(message *commonProto.ConsensusMessage) {
	switch message.Type {
	case commonProto.MessageType_RBC_REQUEST:
	case commonProto.MessageType_RBC_ECHO:
	case commonProto.MessageType_RBC_READY:
	default:
		r.logger.Warningf("Invalid message type: code %d", message.Type)
		return
	}
}

func (r *reliableImpl) sendRequest() {

}

func (r *reliableImpl) recvRequest(request *commonProto.Request) {

}

func (r *reliableImpl) sendEcho() {

}

func (r *reliableImpl) recvEcho(echo *commonProto.Echo) {

}

func (r *reliableImpl) sendReady() {

}

func (r *reliableImpl) recvReady(ready *commonProto.Ready) {

}
