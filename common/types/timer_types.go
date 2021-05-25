package types

import "time"

const (
	BinaryTagTimer = "binary tag timer"
	TxPoolTimer = "transaction pool timer"
)

const (
	DefaultBinaryTagTimer = 5 * time.Second
	DefaultTxPoolTimer = 500 * time.Millisecond
)

const (
	TimeoutBinaryTag = iota
	TimeoutTxPool
)

type TimeoutEvent struct {
	EventType int
	Event     interface{}
}