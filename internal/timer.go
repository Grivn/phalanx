package internal

type Timer interface {
	StartTimer(name string, event interface{})

	StopTimer(name string)

	ClearTimer()
}
