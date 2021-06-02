package event

import "github.com/google/btree"

const (
	BTreeEventPreOrder = iota
	BTreeEventOrder
)

type BtreeEvent struct {
	EventType int
	Seq    uint64
	Digest string
	Event  interface{}
}

func (event *BtreeEvent) Less(item btree.Item) bool {
	return event.Seq < (item.(*BtreeEvent)).Seq
}
