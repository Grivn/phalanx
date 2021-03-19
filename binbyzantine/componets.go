package binbyzantine

import commonProto "github.com/Grivn/phalanx/common/types/protos"

type cert struct {
	finished bool

	counter map[string]map[uint64]bool

	tags map[string]*commonProto.BinaryTag

	bits map[uint64]uint64
}

type qcCert struct {
	finished bool

	broadcast bool

	counter map[string]map[uint64]bool

	tags map[string]*commonProto.BinaryTag
}
