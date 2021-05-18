package types

import "time"

const (
	BinaryTagTimer = "binary_tag_timer"
	TxPoolTimer = "txpool_timer"
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