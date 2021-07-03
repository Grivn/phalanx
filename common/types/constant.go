package types

import "time"

const (
	// DefaultTimeDuration is the default time duration for proposal generation.
	DefaultTimeDuration = 100 * time.Millisecond

	// DefaultLogRotation is the default log rotation for proposal generation.
	DefaultLogRotation int = 50
)
