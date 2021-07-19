package internal

import "github.com/Grivn/phalanx/common/types"

type Executor interface {
	Run()

	// Commit is used to commit the partial orders.
	Commit(event *types.CommitEvent)
}
