package types

import (
	"github.com/Grivn/phalanx/external"
)

type Config struct {
	Author uint64

	Size int

	ReplyC chan ReplyEvent

	Network external.Network
	Logger  external.Logger
}

const (
	RecvRecordTxEvent = iota
	RecvRecordBatchEvent
	RecvLoadBatchEvent
)

type RecvEvent struct {
	EventType int
	Event     interface{}
}

const (
	ReplyBatchEvent = iota
	ReplyMissingBatchEvent
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}
