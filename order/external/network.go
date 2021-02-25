package external

type Network interface {
	Broadcast(msg interface{})
}
