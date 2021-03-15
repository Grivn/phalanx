package types

const (
	ReqReplyBatchByOrder = iota
	ReqReplyProposedRequest
)

type ReplyEvent struct {
	EventType int
	Event     interface{}
}
