package types

import "time"

const (
	BinaryTagTimer = "binary_tag_timer"
)

const (
	DefaultBinaryTagTimer = 5 * time.Second
)

const (
	TimeoutBinaryTag = iota
)

type TimeoutEvent struct {
	EventType int
	Event     interface{}
}
