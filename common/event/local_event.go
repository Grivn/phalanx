package event

type LocalEventType int

const (
	LocalEventCommand LocalEventType = iota
	LocalEventConsensusMessage
)

type LocalEvent struct {
	Type  LocalEventType
	Event interface{}
}
