package internal

import "github.com/Grivn/phalanx/common/protos"

type TestReceiver interface {
	ProcessTransaction(tx *protos.PTransaction)
}
