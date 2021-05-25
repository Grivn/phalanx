package types

import (
	"github.com/gogo/protobuf/sortkeys"
)

type LogID struct {
	Author uint64
	Hash   string
}

type ExecuteLog struct {
	ID string
	N  int
	F  int
	Timestamps []int64
	TrustedTs int64
}

func NewLog(n int,id string) *ExecuteLog {
	return &ExecuteLog{
		ID: id,
		N:  n,
		F:  (n-1)/3,
		Timestamps: nil,
	}
}

func (l *ExecuteLog) Len() int {
	return len(l.Timestamps)
}

func (l *ExecuteLog) Quorum() int {
	return l.N - l.F
}

func (l *ExecuteLog) Update(timestamp int64) {
	l.Timestamps = append(l.Timestamps, timestamp)
}

func (l *ExecuteLog) IsQuorum() bool {
	return l.Len() >= l.Quorum()
}

func (l *ExecuteLog) TrustedTimestamp() int64 {
	if l.TrustedTs != 0 {
		return l.TrustedTs
	}
	sortkeys.Int64s(l.Timestamps)
	l.TrustedTs = l.Timestamps[l.F]
	return l.TrustedTs
}