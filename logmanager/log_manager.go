package logmanager

import "github.com/Grivn/phalanx/external"

type logManager struct {
	// logger is used to print logs
	logger external.Logger
}

func NewLogManager() *logManager {
	return &logManager{}
}
