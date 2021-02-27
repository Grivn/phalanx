package utils

import (
	"os"

	"github.com/Grivn/phalanx/external"

	"github.com/ultramesh/fancylogger"
)

// NewRawLoggerFile create log file for local cluster tests
func NewRawLoggerFile(hostname string) external.Logger {
	rawLogger := fancylogger.NewLogger("test", fancylogger.DEBUG)

	consoleFormatter := &fancylogger.StringFormatter{
		EnableColors:    true,
		TimestampFormat: "2006-01-02T15:04:05.000",
		IsTerminal:      true,
	}

	//test with logger files
	_ = os.Mkdir("testLogger", os.ModePerm)
	fileName := "testLogger/" + hostname + ".log"
	f, _ := os.Create(fileName)
	consoleBackend := fancylogger.NewIOBackend(consoleFormatter, f)

	rawLogger.SetBackends(consoleBackend)
	rawLogger.SetEnableCaller(true)

	return rawLogger
}

// NewRawLogger create log file for local cluster tests
func NewRawLogger() external.Logger {
	rawLogger := fancylogger.NewLogger("test", fancylogger.DEBUG)

	consoleFormatter := &fancylogger.StringFormatter{
		EnableColors:    true,
		TimestampFormat: "2006-01-02T15:04:05.000",
		IsTerminal:      true,
	}

	consoleBackend := fancylogger.NewIOBackend(consoleFormatter, os.Stdout)

	rawLogger.SetBackends(consoleBackend)
	rawLogger.SetEnableCaller(true)

	return rawLogger
}
