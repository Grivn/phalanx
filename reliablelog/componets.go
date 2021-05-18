package reliablelog

import commonTypes "github.com/Grivn/phalanx/common/types"

type logContent struct {
	author uint64

	sequence uint64

	id commonTypes.LogID
}
