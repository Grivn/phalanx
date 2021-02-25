package collector

import "github.com/Grivn/phalanx/order/requester/types"

func NewCollector(c types.Config) *collectorImpl {
	return newCollectorImpl(c)
}

func (ci *collectorImpl) Start() {
	ci.start()
}

func (ci *collectorImpl) Stop() {
	ci.stop()
}

func (ci *collectorImpl) Reset() {
	return
}
