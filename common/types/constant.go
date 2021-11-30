package types

import (
	"time"
)

const (
	// DefaultTimeDuration is the default time duration for proposal generation.
	DefaultTimeDuration = 2 * time.Millisecond

	// DefaultLogRotation is the default log rotation for proposal generation.
	DefaultLogRotation int = 10000

	// SingleCommandSize is used for phalanx test network for size 1.
	SingleCommandSize int = 1

	// DefaultCommandSize is the default batch size for command transactions.
	DefaultCommandSize int = 500

	// DefaultMulti is the default multi proposers for tx manager.
	DefaultMulti int = 1
)
