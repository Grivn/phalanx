package types

type LocalEventType int

const (
	LocalEventTxs LocalEventType = iota
	LocalEventCommand
	LocalEventCommandIndex
	LocalEventOrderAttempt
)

// LocalEvent is the events from local modules.
type LocalEvent struct {
	Type  LocalEventType
	Event interface{}
}
