package internal

import (
	"github.com/Grivn/phalanx/common/types"
)

type Executor interface {
	// CommitStream is used to commit the partial order stream.
	CommitStream(qStream types.QueryStream) error
}
