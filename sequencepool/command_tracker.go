package sequencepool

import "github.com/Grivn/phalanx/common/types"

type CommandTracker interface {
	Add(digest string)

	Del(digest string)

	NonQuorum(digest string) bool

	IsQuorum(digest string) bool
}

type commandTracker struct {
	quorum int

	proposedCmd map[string]int
}

func NewCommandTracker(n int) *commandTracker {
	return &commandTracker{quorum: types.CalculateQuorum(n), proposedCmd: make(map[string]int)}
}

func (ct *commandTracker) Add(digest string) {
	ct.proposedCmd[digest]++
}

func (ct *commandTracker) Del(digest string) {
	ct.proposedCmd[digest]--
}

func (ct *commandTracker) NonQuorum(digest string) bool {
	count, ok := ct.proposedCmd[digest]

	if !ok {
		return true
	}

	return count < ct.quorum
}

func (ct *commandTracker) IsQuorum(digest string) bool {
	count, ok := ct.proposedCmd[digest]

	if !ok {
		return false
	}

	return count >= ct.quorum
}
