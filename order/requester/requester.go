package requester

import "github.com/Grivn/phalanx/order/requester/types"

func NewRequester(c types.Config) *requesterImpl {
	return newRequesterImpl(c)
}

func (ri *requesterImpl) Start() {
	ri.start()
}

func (ri *requesterImpl) Stop() {
	ri.stop()
}

func (ri *requesterImpl) Reset() {
	return
}
