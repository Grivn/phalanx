package external

type External interface {
	CryptoService
	Executor
	Sender
	Logger
}
