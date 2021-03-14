package phalanx

import (
	"os"

	"github.com/Grivn/phalanx/api"
	authenticator "github.com/Grivn/phalanx/authentication"
	commonProto "github.com/Grivn/phalanx/common/types/protos"
	"github.com/Grivn/phalanx/external"
	"github.com/a8m/envsubst"
	"github.com/spf13/viper"
)

type logMgrImpl struct {
	author uint64

	auth api.Authenticator

	recvC chan interface{}

	closeC chan bool

	logger external.Logger
}

func newLogMgrImpl(author uint64, logger external.Logger) *logMgrImpl {
	usigEnclaveFile, err := envsubst.String(viper.GetString("usig.enclaveFile"))
	if err != nil {
		logger.Errorf("failed to parse USIG enclave filename: %s", err)
		return nil
	}

	keysFile, err := os.Open(viper.GetString("keys"))
	if err != nil {
		logger.Errorf("failed to open keyset file: %s", err)
		return nil
	}

	auth, err := authenticator.NewWithSGXUSIG([]api.AuthenticationRole{api.USIGAuthen}, uint32(author), keysFile, usigEnclaveFile)
	return &logMgrImpl{
		author: author,
		auth:   auth,
		logger: logger,
	}
}

func (lm *logMgrImpl) start() {

}

func (lm *logMgrImpl) stop() {

}

func (lm *logMgrImpl) propose(event interface{}) {
	lm.recvC <- event
}

func (lm *logMgrImpl) listener() {
	for {
		select {
		case <-lm.closeC:
			lm.logger.Notice("exist log listener")
			return
		case ev := <-lm.recvC:
			lm.processLogEvents(ev)
		}
	}
}

func (lm *logMgrImpl) processLogEvents(event interface{}) {
	switch e := event.(type) {
	case *commonProto.BatchId:
		lm.generateLog(e)
	case *commonProto.OrderedMsg:

	}
}

func (lm *logMgrImpl) generateLog(bid *commonProto.BatchId) {

}

func (lm *logMgrImpl) processOrderedMsg(msg *commonProto.OrderedMsg) {
	lm.logger.Infof("Replica %d received ordered log from replica %d in sequence %d for hash %s", lm.author, msg.Author, msg.BatchId.BatchHash)
}
