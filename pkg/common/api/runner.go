package api

// Runner is used to control the modules which provide coroutine service.
type Runner interface {
	Run()
	Quit()
}
