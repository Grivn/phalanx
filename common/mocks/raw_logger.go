package mocks

import (
	"io"
	"os"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// mockLogger is an implement for Logger.
type mockLogger struct {
	logger *logrus.Logger
}

// NewRawLogger provides a Logger instance to print logs in stdout.
func NewRawLogger() *mockLogger {
	log := logrus.New()
	writer := os.Stdout
	log.SetFormatter(&nested.Formatter{NoColors: true})
	log.SetOutput(io.MultiWriter(writer))
	return &mockLogger{logger: log}
}

// NewRawLoggerFile provides a Logger instance to print logs in files.
func NewRawLoggerFile(hostname string) *mockLogger {
	log := logrus.New()
	writer, err := os.OpenFile(hostname+"_"+time.Now().Format("2006-01-02_15:04:05")+".log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&nested.Formatter{NoColors: true})
	log.SetOutput(io.MultiWriter(writer))
	return &mockLogger{logger: log}
}

func (log *mockLogger) Debug(v ...interface{}) {
	log.logger.Debug(v...)
}
func (log *mockLogger) Debugf(format string, v ...interface{}) {
	log.logger.Debugf(format, v...)
}
func (log *mockLogger) Info(v ...interface{}) {
	log.logger.Info(v...)
}
func (log *mockLogger) Infof(format string, v ...interface{}) {
	log.logger.Infof(format, v...)
}
func (log *mockLogger) Error(v ...interface{}) {
	log.logger.Error(v...)
}
func (log *mockLogger) Errorf(format string, v ...interface{}) {
	log.logger.Errorf(format, v...)
}