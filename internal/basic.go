package internal

// Basic is the basic controller for sub-module
type Basic interface {
	// Start is used to start the sub-module
	Start()

	// Stop is used to stop the sub-module
	Stop()
}
