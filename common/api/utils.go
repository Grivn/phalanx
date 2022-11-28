package api

type LocalTimer interface {
	// StartTimer starts current timer.
	StartTimer()

	// StopTimer stops current timer.
	StopTimer()
}
