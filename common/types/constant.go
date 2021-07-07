package types

import "time"

const (
	// DefaultTimeDuration is the default time duration for proposal generation.
	DefaultTimeDuration = 2 * time.Millisecond

	// DefaultLogRotation is the default log rotation for proposal generation.
	DefaultLogRotation int = 50

	// TestBatchSize is used for phalanx test network
	TestBatchSize int = 1
)
