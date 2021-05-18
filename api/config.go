package api

import "time"

//======= Interface for module 'config' =======

// Configer defines the interface to obtain the protocol parameters from
// the configuration
type Configer interface {
	// n: number of nodes in the network
	N() uint32
	// f: number of byzantine nodes the network can tolerate
	F() uint32

	// cp: checkpoint period
	CheckpointPeriod() uint32
	// L: must be larger than CheckpointPeriod
	Logsize() uint32

	// starts when receives a request and stops when request is accepted
	TimeoutRequest() time.Duration

	// starts when receives a request and stops when request is prepared
	TimeoutPrepare() time.Duration

	// starts when sends VIEW-CHANGE and stops when receives a valid NEW-VIEW
	TimeoutViewChange() time.Duration
}
